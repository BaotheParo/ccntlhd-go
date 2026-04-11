package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
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

	// Tao mat khau bam
	hashedPassword, _ := auth.HashPassword("password123")

	email := "user@example.com"
	var existing entity.User
	result := db.Where("email = ?", email).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		newUser := entity.User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         entity.RoleUser,
			IsActive:     true,
		}
		db.Create(&newUser)
		fmt.Printf("✅ Da tao thanh cong: %s voi mat khau: password123\n", email)
	} else {
		// Cap nhat mat khau neu user da ton tai nhung sai pass
		existing.PasswordHash = hashedPassword
		db.Save(&existing)
		fmt.Printf("ℹ️ Da cap nhat mat khau cho: %s thanh: password123\n", email)
	}
}
