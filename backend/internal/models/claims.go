package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// JWT Claims struct
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}
