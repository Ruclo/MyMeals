package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Order struct {
	ID         uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	TableNo    int         `gorm:"check:table_no >= 1" json:"table_no" binding:"required"`
	Notes      string      `gorm:"not null" json:"notes"`
	OrderMeals []OrderMeal `gorm:"foreignKey:OrderID; preload:true" json:"order_meals" binding:"required"`
	CreatedAt  time.Time   `json:"created_at"`
	Review     *Review     `gorm:"foreignKey:OrderID" json:"review,omitempty"`
}

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
	OrderID   uint  `gorm:"primaryKey" json:"order_id"`
	MealID    uint  `gorm:"primaryKey" json:"meal_id" binding:"required"`
	Quantity  uint  `json:"quantity" binding:"required"`
	Completed uint  `json:"completed"`
	Meal      *Meal `gorm:"foreignKey:MealID; preload:true" json:"-"`
}

func (om *OrderMeal) BeforeCreate(tx *gorm.DB) error {
	//om.Completed = 0
	return nil
}
