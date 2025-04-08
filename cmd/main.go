package main

import (
	"split-the-bill/internal/config"
	"split-the-bill/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.InitDB()
	r := gin.Default()
	routes.SetupRoutes(r, db)
	r.Run(":8080")
}
