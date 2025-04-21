package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"split-the-bill/internal/controllers"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.POST("/users", controllers.CreateUser(db))
	r.GET("/users", controllers.ListUsers(db))

	r.POST("/events", controllers.CreateEvent(db))
	r.GET("/events/:id", controllers.GetEvent(db))
	r.POST("/events/:id/participants", controllers.AddParticipant(db))
	r.GET("/events/:id/participants", controllers.ListParticipants(db))

	r.POST("/events/:id/expenses", func(c *gin.Context) {
		controllers.AddExpense(c, db)
	})
	r.GET("/events/:id/expenses", controllers.ListExpenses(db))
	r.PUT("/expenses/:id", controllers.UpdateExpense(db))
	r.DELETE("/expenses/:id", controllers.DeleteExpense(db))

	r.GET("/events/:id/summary", controllers.GetEventSummary(db))

	r.GET("/events/:id/debts", controllers.GetDebts(db))
	r.POST("/events/:id/payments", func(c *gin.Context) {
		controllers.AddPayment(c, db)
	})
	r.GET("/events/:id/payments", controllers.ListPayments(db))
}
