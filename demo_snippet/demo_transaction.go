package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Kết nối DB (Dùng cổng 5433 của máy host để gọi vào Docker)
	dsn := "host=localhost user=user password=password dbname=ticket_db port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Ket noi DB that bai: %v\n", err)
		return
	}

	// 2. Bat dau thực thi Giao dịch (Transaction)
	tx := db.Begin()
	fmt.Println("Bat dau thuc thi Giao dich (Transaction)...")

	// Bước A: Tạo sự kiện mới (Thành công)
	eventID := uuid.New()
	event := entity.Event{
		ID:        eventID,
		Name:      "Su kien kiem thu Transaction",
		Slug:      "su-kien-kiem-thu",
		Location:  "Phong Lab",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Hour),
		Status:    entity.EventStatusDraft,
	}

	if err := tx.Create(&event).Error; err != nil {
		fmt.Printf("Buoc A that bai: %v\n", err)
		tx.Rollback()
		return
	}
	fmt.Println("Thao tac 1 thanh cong: Da khoi tao Su kien trong bo nho tam.")

	// Bước B: Cố tình tạo một Sự kiện khác nhưng TRÙNG SLUG
	// Slug "su-kien-kiem-thu" đã được dùng bên trên, gây lỗi Duplicate
	errorEvent := entity.Event{
		ID:        uuid.New(),
		Name:      "Su kien bi loi",
		Slug:      "su-kien-kiem-thu", // TRÙNG SLUG ĐÃ CÓ
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}

	fmt.Println("Dang tien hanh Thao tac 2 (Cu co tinh tao trung Slug)...")
	if err := tx.Create(&errorEvent).Error; err != nil {
		fmt.Println("Thao tac 2 THAT BAI: Du lieu khong hop le (Duplicate unique slug)!")
		fmt.Println("Kich hoat co che ROLLBACK... Huy bo toan bo Giao dich!")
		tx.Rollback()
		return
	}

	// Neu may man vuot qua (nhung se khong the)
	tx.Commit()
	fmt.Println("Giao dich hoan tat thanh cong!")
}
