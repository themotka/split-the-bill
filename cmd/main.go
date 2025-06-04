package main

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
	"split-the-bill/internal/config"
	"split-the-bill/internal/controllers"
	"split-the-bill/internal/middleware"
	"split-the-bill/internal/routes"
)

func main() {
	db := config.InitDB()
	r := gin.Default()

	log := setupLogger()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Error("JWT_SECRET is not set")
		os.Exit(1)
	}

	r.POST("/login", controllers.LoginHandler(db, jwtSecret))
	r.POST("/register", controllers.RegisterHandler(db, jwtSecret))

	authorized := r.Group("/")

	authorized.Use(middleware.AuthMiddleware(jwtSecret, db, log))

	routes.SetupRoutes(authorized, db)
	err := r.Run(":8080")
	if err != nil {
		log.Error("Error starting server")
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	return log
}
