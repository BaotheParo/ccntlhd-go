package main

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	jwtSecret := []byte("my-super-secret-key-2026")
	
	claims := jwt.MapClaims{
		"user_id": userID,
		"sub":     userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	
	fmt.Println(tokenString)
}
