package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/core/entity"
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
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Khong the ket noi DB: %v", err)
	}

	// 1. Tim User
	var user entity.User
	db.Where("email = ?", "user@example.com").First(&user)
	if user.ID == uuid.Nil {
		log.Fatal("Khong tim thay user@example.com. Vui long chay seed_user truoc.")
	}

	// 2. Tim Ticket Type (Lay hang VIP)
	var tt entity.TicketType
	db.Where("name = ?", "VIP").First(&tt)
	if tt.ID == uuid.Nil {
		log.Fatal("Khong tim thay hang ve VIP.")
	}

	// 3. Tao Đon hang gia (Mock Order)
	orderID := uuid.New()
	quantity := 1
	totalAmount := tt.Price.Mul(decimal.NewFromInt(int64(quantity)))

	order := entity.Order{
		ID:          orderID,
		UserID:      user.ID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPaid,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	orderItem := entity.OrderItem{
		ID:           uuid.New(),
		OrderID:      orderID,
		TicketTypeID: tt.ID,
		Quantity:     quantity,
		UnitPrice:    tt.Price,
	}

	// 4. Luu vao DB
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Loi khi tao don hang gia: %v", err)
	}

	fmt.Printf("✅ Da tao don hang thanh cong cho %s (OrderID: %s)\n", user.Email, orderID)
}
