package main

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Usernames  []Username
	Purchases  []Purchase
	Payments   []Payment
	IsDisabled *bool `gorm:"default:false"`
}

type Username struct {
	gorm.Model
	Name   string
	UserID uint64
	User   User
}

type Product struct {
	gorm.Model
	Name       string
	Price      uint64
	Barcode    string
	IsDisabled *bool `gorm:"default:false"`
}

type Purchase struct {
	gorm.Model
	UserID          uint64
	User            User
	PurchaseDetails []PurchaseDetail
	PaymentID       *uint64
	Payment         *Payment
}

type PurchaseDetail struct {
	gorm.Model
	Quantity   uint64
	Total      uint64
	ProductID  uint64
	Product    Product
	PurchaseID uint64
	Purchase   Purchase
}

type Payment struct {
	gorm.Model
	UserID    uint64
	User      User
	Purchases []Purchase
}

type BuyRequestDetail struct {
	ProductID uint64
	Quantity  uint64
}

type BuyRequest struct {
	UserID  uint64
	Details []BuyRequestDetail
}

type OldSystemTransaction struct {
	CreatedAt float64  `json:"created_at"`
	DeletedAt *float64 `json:"deleted_at"`
	Action    string   `json:"action"`
	Amount    uint64   `json:"amount"`
}

type OldSystemData struct {
	User         string                 `json:"user"`
	CreatedAt    int64                  `json:"created_at"`
	Transactions []OldSystemTransaction `json:"transactions"`
}
