package main

import (
	"fmt"
	"log"

	"github.com/yourname/ticketing-system/internal/core/entity"
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

	var users []entity.User
	db.Find(&users)
	fmt.Println("--- USERS ---")
	for _, u := range users {
		fmt.Printf("ID: %s, Email: %s, Role: %s\n", u.ID, u.Email, u.Role)
	}

	var orders []entity.Order
	db.Find(&orders)
	fmt.Println("\n--- ORDERS ---")
	for _, o := range orders {
		fmt.Printf("ID: %s, UserID: %s, Status: %s, Amount: %v\n", o.ID, o.UserID, o.Status, o.TotalAmount)
	}

	var ticketTypes []entity.TicketType
	db.Find(&ticketTypes)
	fmt.Println("\n--- TICKET TYPES ---")
	for _, t := range ticketTypes {
		fmt.Printf("ID: %s, Name: %s, Remaining: %d\n", t.ID, t.Name, t.RemainingQuantity)
	}
}
