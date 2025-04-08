package models

import "time"

type User struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	AvatarURL *string
	Email     *string `gorm:"unique"`
	CreatedAt time.Time
}

type Event struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CreatedBy uint
	CreatedAt time.Time
}

type EventParticipant struct {
	ID      uint `gorm:"primaryKey"`
	EventID uint
	UserID  uint
}

type Expense struct {
	ID      uint `gorm:"primaryKey"`
	EventID uint
	Title   string
	Amount  float64
	PaidBy  uint
	PaidAt  time.Time
}

type ExpenseShare struct {
	ID          uint `gorm:"primaryKey"`
	ExpenseID   uint
	UserID      uint
	ShareAmount float64
}

type Debt struct {
	ID        uint `gorm:"primaryKey"`
	EventID   uint
	FromUser  uint
	ToUser    uint
	Amount    float64
	IsSettled bool
}

type Payment struct {
	ID       uint `gorm:"primaryKey"`
	FromUser uint
	ToUser   uint
	Amount   float64
	PaidAt   time.Time
	EventID  uint
}
