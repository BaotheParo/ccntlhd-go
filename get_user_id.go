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

	var u entity.User
	db.Where("email = ?", "user@example.com").First(&u)
	fmt.Printf("User: %s, ID: %s\n", u.Email, u.ID)
}
