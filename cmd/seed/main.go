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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Khong the tai cau hinh: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Khong the ket noi DB: %v", err)
	}

	// Tao mat khau bam cho 'admin'
	hashedPassword, _ := auth.HashPassword("admin")

	admin := entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		Role:         entity.RoleAdmin,
		IsActive:     true,
	}

	// Kiem tra xem user da ton tai chua
	var existing entity.User
	result := db.Where("email = ?", "admin@example.com").First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("Loi khi tao admin: %v", err)
		}
		fmt.Println("✅ Da tao tai khoan admin@example.com thanh cong (mat khau: admin)")
	} else {
		fmt.Println("ℹ️ Tai khoan admin@example.com da ton tai.")
	}
}
