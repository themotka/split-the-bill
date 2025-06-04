package main

import (
	"context"
	"log/slog"
	"os"
	"split-the-bill/internal/clients"
	"split-the-bill/internal/config"
	"split-the-bill/internal/controllers"
	"split-the-bill/internal/middleware"
	"split-the-bill/internal/routes"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.InitDB()
	r := gin.Default()

	log := setupLogger()

	jwtSecret := os.Getenv("JWT_SECRET")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ssoClient, err := clients.New(ctx, log,
		"127.0.0.1:5433",
		5,
		2)

	if err != nil {
		log.Error("Error creating sso client")
		os.Exit(1)
	}

	r.POST("/login", controllers.LoginHandler(ssoClient, log))
	r.POST("/register", controllers.RegisterHandler(ssoClient, log))

	authorized := r.Group("/")

	authorized.Use(middleware.AuthMiddleware(jwtSecret, db, log))

	routes.SetupRoutes(authorized, db)
	err = r.Run(":8080")
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
