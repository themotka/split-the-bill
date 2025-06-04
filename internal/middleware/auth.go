package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"log/slog"
	"math/rand"
	"net/http"
	"split-the-bill/internal/models"
	"strings"
)

func AuthMiddleware(secret string, db *gorm.DB, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			log.Error("Missing or invalid Authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Error("Unexpected signing method")
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			log.Error("Invalid token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			log.Error("Invalid claims")
			c.Abort()
			return
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UID missing or invalid"})
			log.Error("UID missing or invalid")
			c.Abort()
			return
		}
		uid := uint(uidFloat)

		email, _ := claims["email"].(string)

		var user models.User
		if err := db.First(&user, uid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Попробуем найти по email (на случай, если uid другой)
				if email != "" {
					err = db.Where("email = ?", email).First(&user).Error
				}
			}
		}
		avatar := fmt.Sprintf("https://1avatara.ru/pic/animal/animal000%d.jpg", rand.Intn(9)+1)
		if user.ID == 0 {
			// Пользователь не найден, создаём нового
			user = models.User{
				ID:        uid,
				Email:     &email,
				AvatarURL: &avatar,
			}
			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				log.Error("Failed to create user")
				c.Abort()
				return
			}
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}
