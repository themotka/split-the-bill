package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	dbUser     = "postgres"
	dbPassword = "yourpassword"
	dbName     = "split_the_bill"
)

var db *sql.DB

func main() {
	var err error
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.POST("/users", createUser)
	r.POST("/events", createEvent)
	r.POST("/events/:id/participants", addParticipant)
	r.POST("/events/:id/expenses", addExpense)
	r.GET("/events/:id/summary", getEventSummary)

	r.Run(":8080")
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, user.Name, user.Email).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func createEvent(c *gin.Context) {
	type Request struct {
		Name      string `json:"name"`
		CreatedBy int    `json:"created_by"`
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `INSERT INTO events (name, created_by) VALUES ($1, $2) RETURNING id`
	var eventID int
	err := db.QueryRow(query, req.Name, req.CreatedBy).Scan(&eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"event_id": eventID})
}

func addParticipant(c *gin.Context) {
	eventID := c.Param("id")
	type Request struct {
		UserID int `json:"user_id"`
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `INSERT INTO event_participants (event_id, user_id) VALUES ($1, $2)`
	_, err := db.Exec(query, eventID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "participant added"})
}

func addExpense(c *gin.Context) {
	eventID := c.Param("id")
	type Share struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	type Request struct {
		Title  string  `json:"title"`
		Amount float64 `json:"amount"`
		PaidBy int     `json:"paid_by"`
		Shares []Share `json:"shares"`
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var expenseID int
	err = tx.QueryRow(`INSERT INTO expenses (event_id, title, amount, paid_by) VALUES ($1, $2, $3, $4) RETURNING id`, eventID, req.Title, req.Amount, req.PaidBy).Scan(&expenseID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range req.Shares {
		_, err := tx.Exec(`INSERT INTO expense_shares (expense_id, user_id, share_amount) VALUES ($1, $2, $3)`, expenseID, s.UserID, s.Amount)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"expense_id": expenseID})
}

func getEventSummary(c *gin.Context) {
	eventID := c.Param("id")
	rows, err := db.Query(`
		SELECT u.id, u.name, 
			SUM(CASE WHEN e.paid_by = u.id THEN e.amount ELSE 0 END) AS paid,
			SUM(CASE WHEN es.user_id = u.id THEN es.share_amount ELSE 0 END) AS owed
		FROM users u
		LEFT JOIN expenses e ON e.event_id = $1 AND e.paid_by = u.id
		LEFT JOIN expense_shares es ON es.expense_id = e.id AND es.user_id = u.id
		GROUP BY u.id, u.name
	`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	type Summary struct {
		UserID  int     `json:"user_id"`
		Name    string  `json:"name"`
		Paid    float64 `json:"paid"`
		Owed    float64 `json:"owed"`
		Balance float64 `json:"balance"`
	}
	var summary []Summary
	for rows.Next() {
		var s Summary
		if err := rows.Scan(&s.UserID, &s.Name, &s.Paid, &s.Owed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		s.Balance = s.Paid - s.Owed
		summary = append(summary, s)
	}
	c.JSON(http.StatusOK, summary)
}
