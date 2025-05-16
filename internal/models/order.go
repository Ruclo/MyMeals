package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Order struct {
	ID         uint        `gorm:"primaryKey;autoIncrement"`
	TableNo    int         `gorm:"check:table_no >= 1"`
	Notes      string      `gorm:"not null"`
	OrderMeals []OrderMeal `gorm:"foreignKey:OrderID; preload:true"`
	CreatedAt  time.Time
	Review     *Review `gorm:"foreignKey:OrderID"`
}

// BeforeCreate is a GORM hook that validates and resets fields before creating an Order record in the database.
// It ensures the order contains at least one meal and resets the CreatedAt field if it is not zero.
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if len(o.OrderMeals) == 0 {
		return errors.New("order must have at least one meal")
	}

	if !o.CreatedAt.IsZero() {
		o.CreatedAt = time.Time{}
	}
	return nil
}

type OrderMeal struct {
	OrderID   uint `gorm:"primaryKey"`
	MealID    uint `gorm:"primaryKey"`
	Quantity  uint
	Completed uint
	Meal      *Meal `gorm:"foreignKey:MealID"`
}
