package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&User{}, &Username{}, &Product{}, &Purchase{}, &PurchaseDetail{}, &Payment{})

	if err := db.Take(&Product{}, 0).Error; err != nil {
		db.Create(&Product{
			Name:    "其他",
			Price:   1,
			Barcode: "",
		})
	}

	return db
}
