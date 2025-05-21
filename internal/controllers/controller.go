package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"split-the-bill/internal/common"
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

type ShareInput struct {
	UserID      uint    `json:"user_id"`
	ShareAmount float64 `json:"share_amount"`
}

type CreateExpenseInput struct {
	Title  string       `json:"title"`
	Amount float64      `json:"amount"`
	PaidBy uint         `json:"paid_by"`
	PaidAt *time.Time   `json:"paid_at"`
	Shares []ShareInput `json:"shares"`
}

func AddExpense(c *gin.Context, db *gorm.DB) {
	eventID := c.Param("id")
	fmt.Print(eventID)
	var input CreateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var totalShares float64
	for _, share := range input.Shares {
		totalShares += share.ShareAmount
	}
	if totalShares != input.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Сумма долей не совпадает с общей суммой"})
		return
	}

	expense := models.Expense{
		EventID: common.ParseUintParam(eventID),
		Title:   input.Title,
		Amount:  input.Amount,
		PaidBy:  input.PaidBy,
		PaidAt:  time.Now(),
	}

	if input.PaidAt != nil {
		expense.PaidAt = *input.PaidAt
	}

	if err := db.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании траты"})
		return
	}

	for _, s := range input.Shares {
		share := models.ExpenseShare{
			ExpenseID:   expense.ID,
			UserID:      s.UserID,
			ShareAmount: s.ShareAmount,
		}
		db.Create(&share)
	}

	for _, s := range input.Shares {
		if s.UserID == input.PaidBy || s.ShareAmount == 0 {
			continue
		}

		fromID := s.UserID
		toID := input.PaidBy
		amount := s.ShareAmount

		var reverseDebt models.Debt
		err := db.Where("event_id = ? AND from_user = ? AND to_user = ? AND is_settled = false",
			expense.EventID, toID, fromID).First(&reverseDebt).Error

		if err == nil {
			if reverseDebt.Amount > amount {
				reverseDebt.Amount -= amount
				db.Save(&reverseDebt)
			} else if reverseDebt.Amount < amount {
				db.Delete(&reverseDebt)

				newAmount := amount - reverseDebt.Amount
				newDebt := models.Debt{
					EventID:  expense.EventID,
					FromUser: fromID,
					ToUser:   toID,
					Amount:   newAmount,
				}
				db.Create(&newDebt)
			} else {
				db.Delete(&reverseDebt)
			}

		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Нет обратного — ищем прямой
			var existingDebt models.Debt
			err := db.Where("event_id = ? AND from_user = ? AND to_user = ? AND is_settled = false",
				expense.EventID, fromID, toID).First(&existingDebt).Error

			if err == nil {
				existingDebt.Amount += amount
				db.Save(&existingDebt)
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				newDebt := models.Debt{
					EventID:  expense.EventID,
					FromUser: fromID,
					ToUser:   toID,
					Amount:   amount,
				}
				db.Create(&newDebt)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске долга"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске обратного долга"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Трата добавлена", "expense_id": expense.ID})
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

func AddPayment(c *gin.Context, db *gorm.DB) {
	var input models.Payment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var totalDebt float64
		if err := tx.Model(&models.Debt{}).
			Where("event_id = ? AND from_user = ? AND to_user = ? AND is_settled = false",
				input.EventID, input.FromUser, input.ToUser).
			Select("COALESCE(SUM(amount), 0)").Scan(&totalDebt).Error; err != nil {
			return err
		}

		if input.Amount > totalDebt {
			return fmt.Errorf("сумма оплаты превышает текущий долг (%.2f)", totalDebt)
		}

		if err := tx.Create(&input).Error; err != nil {
			return err
		}

		remaining := input.Amount

		var debts []models.Debt
		if err := tx.Where("event_id = ? AND from_user = ? AND to_user = ? AND is_settled = false",
			input.EventID, input.FromUser, input.ToUser).
			Order("id").
			Find(&debts).Error; err != nil {
			return err
		}

		for _, debt := range debts {
			if remaining <= 0 {
				break
			}

			if remaining >= debt.Amount {
				remaining -= debt.Amount
				debt.Amount = 0
				debt.IsSettled = true
			} else {
				debt.Amount -= remaining
				remaining = 0
			}

			if err := tx.Save(&debt).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Платёж успешно добавлен и долги обновлены"})
}

func ListPayments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("id")
		var payments []models.Payment
		db.Where("event_id = ?", eventID).Find(&payments)
		c.JSON(http.StatusOK, payments)
	}
}
