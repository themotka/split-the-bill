package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"split-the-bill/internal/models"
)

func InitDB() *gorm.DB {
	dsn := "host=localhost user=user password=password dbname=splitwise_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.EventParticipant{},
		&models.Expense{},
		&models.ExpenseShare{},
		&models.Debt{},
		&models.Payment{},
	)
	if err != nil {
		log.Fatal("Failed to migrate DB:", err)
		return nil
	}

	return db
}
