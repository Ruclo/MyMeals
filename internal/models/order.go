package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type OrderStatus string

const (
	StatusDone    OrderStatus = "Done"
	StatusPending OrderStatus = "Pending"
)

func (s OrderStatus) Valid() error {
	switch s {
	case StatusDone, StatusPending:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid order status %s", s))
	}
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid scan source for OrderStatus")
		}
		str = string(bytes)
	}

	*s = OrderStatus(str)
	return s.Valid()
}

func (s OrderStatus) Value() (driver.Value, error) {
	if err := s.Valid(); err != nil {
		return nil, err
	}
	return string(s), nil
}

type Order struct {
	ID         uint        `gorm:"primaryKey;autoIncrement"`
	TableNo    int         `gorm:"check:table_no >= 1"`
	Name       string      `gorm:"not null"`
	Notes      string      `gorm:"not null"`
	OrderMeals []OrderMeal `gorm:"foreignKey:OrderID"`
	CreatedAt  time.Time
	Review     *Review `gorm:"foreignKey:OrderID"`
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
	OrderID  uint   `gorm:"primaryKey"`
	MealID   uint   `gorm:"primaryKey"`
	MealName string `gorm:"-"`
	Quantity int    `gorm:"check:quantity >= 1"`
	Status   OrderStatus
	Meal     *Meal `gorm:"foreignKey:MealID"`
}

func (om *OrderMeal) BeforeCreate(tx *gorm.DB) error {
	om.Status = StatusPending
	return nil
}

func (om *OrderMeal) AfterFind(db *gorm.DB) error {
	om.MealName = om.Meal.Name
	return nil
}
