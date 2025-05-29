package main

import (
	"fmt"
	"os"
	"split-the-bill/internal/config"
	"split-the-bill/internal/middleware"
	"split-the-bill/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.InitDB()
	r := gin.Default()

	jwtSecret := os.Getenv("JWT_SECRET")

	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware(jwtSecret, db))

	routes.SetupRoutes(authorized, db)
	err := r.Run(":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
}
