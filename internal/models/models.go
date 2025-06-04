package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
	Email     *string   `gorm:"unique"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"-"`
}

type Event struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type EventParticipant struct {
	ID      uint `gorm:"primaryKey"`
	EventID uint `json:"event_id"`
	UserID  uint `json:"user_id"`
}

type Expense struct {
	ID      uint      `gorm:"primaryKey"`
	EventID uint      `json:"event_id"`
	Title   string    `json:"title"`
	Amount  float64   `json:"amount"`
	PaidBy  uint      `json:"paid_by"`
	PaidAt  time.Time `json:"created_at"`
}

type ExpenseShare struct {
	ID          uint    `gorm:"primaryKey"`
	ExpenseID   uint    `json:"expense_id"`
	UserID      uint    `json:"user_id"`
	ShareAmount float64 `json:"share_amount"`
}

type Debt struct {
	ID        uint    `gorm:"primaryKey"`
	EventID   uint    `json:"event_id"`
	FromUser  uint    `json:"from_user"`
	ToUser    uint    `json:"to_user"`
	Amount    float64 `json:"amount"`
	IsSettled bool    `json:"is_settled"`
}

type Payment struct {
	ID       uint      `gorm:"primaryKey"`
	FromUser uint      `json:"from_user"`
	ToUser   uint      `json:"to_user"`
	Amount   float64   `json:"amount"`
	PaidAt   time.Time `json:"created_at"`
	EventID  uint      `json:"event_id"`
}
