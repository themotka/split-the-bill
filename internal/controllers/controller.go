package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"split-the-bill/internal/models"
	"strconv"
	"time"
)

func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&user)
		c.JSON(http.StatusOK, user)
	}
}

func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	}
}

func CreateEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var event models.Event
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&event)
		c.JSON(http.StatusOK, event)
		var p models.EventParticipant

		p.UserID = event.CreatedBy
		p.EventID = event.ID
		db.Create(&p)
	}
}

func GetEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var event models.Event
		if err := db.First(&event, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusOK, event)
	}
}

func AddParticipant(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p models.EventParticipant
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		p.EventID = uint(id)
		db.Create(&p)
		c.JSON(http.StatusOK, p)
	}
}

func ListParticipants(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var participants []models.EventParticipant
		db.Where("event_id = ?", eventID).Find(&participants)
		c.JSON(http.StatusOK, participants)
	}
}

func AddExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var expense models.Expense
		if err := c.ShouldBindJSON(&expense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		expense.EventID = uint(id)
		if expense.PaidAt.IsZero() {
			expense.PaidAt = time.Now()
		}
		db.Create(&expense)

		var shares []models.ExpenseShare
		if err := c.ShouldBindJSON(&shares); err == nil {
			for i := range shares {
				shares[i].ExpenseID = expense.ID
			}
			db.Create(&shares)
		}

		c.JSON(http.StatusOK, expense)
	}
}

func ListExpenses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var expenses []models.Expense
		db.Where("event_id = ?", eventID).Find(&expenses)
		c.JSON(http.StatusOK, expenses)
	}
}

func UpdateExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var expense models.Expense
		if err := db.First(&expense, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
			return
		}
		if err := c.ShouldBindJSON(&expense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Save(&expense)
		c.JSON(http.StatusOK, expense)
	}
}

func DeleteExpense(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Expense{}, id)
		db.Where("expense_id = ?", id).Delete(&models.ExpenseShare{})
		c.Status(http.StatusNoContent)
	}
}

func GetEventSummary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var total float64
		db.Model(&models.Expense{}).Where("event_id = ?", eventID).Select("SUM(amount)").Scan(&total)

		type Share struct {
			UserID uint    `json:"user_id"`
			Amount float64 `json:"amount"`
		}

		var shares []Share
		db.Table("expense_shares").Select("user_id, SUM(share_amount) as amount").
			Joins("JOIN expenses ON expense_shares.expense_id = expenses.id").
			Where("expenses.event_id = ?", eventID).
			Group("user_id").Scan(&shares)

		c.JSON(http.StatusOK, gin.H{"total": total, "shares": shares})
	}
}

func GetDebts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var debts []models.Debt
		db.Where("event_id = ? AND is_settled = false", eventID).Find(&debts)
		c.JSON(http.StatusOK, debts)
	}
}

func RecordPayment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payment models.Payment
		if err := c.ShouldBindJSON(&payment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		payment.EventID = uint(id)
		payment.PaidAt = time.Now()
		db.Create(&payment)
		c.JSON(http.StatusOK, payment)
	}
}

func ListPayments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var payments []models.Payment
		db.Where("event_id = ?", eventID).Find(&payments)
		c.JSON(http.StatusOK, payments)
	}
}
