package models

import "time"

type Review struct {
	ID        uint   `gorm:"primaryKey"`
	OrderID   uint   `gorm:"unique; not null; constraint: OnDelete:CASCADE, OnUpdate:CASCADE; references:orders(ID)"`
	Rating    int    `gorm:"check:rating >= 1 AND rating <= 5"`
	Comment   string `gorm:"not null; check:comment <>"`
	CreatedAt time.Time
	PhotoURLs []string `gorm:"type:text[]; check: (SELECT COUNT(*) FROM unnest(photo_urls) url WHERE length(url) = 0) = 0"`
}
