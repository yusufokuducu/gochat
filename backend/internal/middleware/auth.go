package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

// Auth is the authentication middleware
type Auth struct {
	JWTSecret []byte
}

// NewAuth creates a new authentication middleware
func NewAuth(jwtSecret []byte) *Auth {
	return &Auth{
		JWTSecret: jwtSecret,
	}
}

// Middleware returns the authentication middleware
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Extract the token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected 'Bearer TOKEN'"})
			return
		}

		tokenString := authHeader[7:] // Remove "Bearer " prefix
		claims := &models.Claims{}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return a.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set user ID in context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
