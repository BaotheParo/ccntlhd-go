package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/pkg/auth"
	"github.com/yourname/ticketing-system/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, _ := config.LoadConfig()
	
	// Khởi tạo DB
	dsn := fmt.Sprintf("host=localhost port=5433 user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Lỗi connect DB: ", err)
	}

	// 1. Lấy thông tin User và Gen Token
	var user entity.User
	db.Where("email = ?", "user@example.com").First(&user)
	if user.ID.String() == "00000000-0000-0000-0000-000000000000" {
		log.Fatal("Chưa có User nào trong hệ thống, hãy tạo user trước!")
	}
	secret := "my-super-secret-key-2026"
	token, _ := auth.GenerateToken(user.ID.String(), user.Role, secret)

	// 2. Lấy 1 Ticket Type từ DB để test Flash Sale
	var ticket entity.TicketType
	if err := db.First(&ticket).Error; err != nil {
		log.Fatal("Không tìm thấy Loại vé (Ticket Type) nào trong hệ thống. Vui lòng tạo Event & TicketType trước.")
	}

	// 3. Khởi tạo/Cập nhật số lượng vé vào Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
	redisKey := fmt.Sprintf("ticket:%s:stock", ticket.ID.String())
	
	// Mở kho 10.000 vé cho Load Test bung nóc
	err = rdb.Set(context.Background(), redisKey, 10000, 0).Err()
	if err != nil {
		log.Fatal("Lỗi kết nối Redis: ", err)
	}

	fmt.Println("\n=======================================================")
	fmt.Println("🚀 SETUP HOÀN TẤT: Đã bơm 10.000 vé vào Redis cho ID:", ticket.ID)
	fmt.Println("=======================================================\n")
	fmt.Println("Hãy copy và chạy lệnh K6 này:\n")
	fmt.Printf("k6 run -e BASE_URL=http://localhost:8080/api/v1 -e TOKEN=\"%s\" -e TICKET_ID=\"%s\" tests/load/flash_sale_test.js\n\n", token, ticket.ID.String())
}
