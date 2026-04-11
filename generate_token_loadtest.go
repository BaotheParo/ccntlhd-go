package main

import (
	"fmt"
	"log"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/pkg/auth"
	"github.com/yourname/ticketing-system/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, _ := config.LoadConfig()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Tim User
	var user entity.User
	db.Where("email = ?", "user@example.com").First(&user)
	if user.ID.String() == "00000000-0000-0000-0000-000000000000" {
		log.Fatal("User not found")
	}

	// 2. Lay Secret (Mac dinh trong main.go)
	secret := "my-super-secret-key-2026" 

	// 3. Tao Token
	token, err := auth.GenerateToken(user.ID.String(), user.Role, secret)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("FRESH_TOKEN_START")
	fmt.Println(token)
	fmt.Println("FRESH_TOKEN_END")
}
