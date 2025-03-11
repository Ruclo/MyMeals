package models

import "time"

type Review struct {
	ID        uint      `gorm:"primaryKey"`
	OrderID   uint      `gorm:"unique"`
	Rating    int       `gorm:"check:rating >= 1 AND rating <= 5"`
	ReviewText string
	CreatedAt     time.Time
	PhotoURLs    []string `gorm:"type:text[]"`
}