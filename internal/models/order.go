package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
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
	ID      uint    `gorm:"primaryKey"`
	TableNo int     `gorm:"check:table_no >= 1"`
	Name    string  `gorm:"not null; check name <> ''"`
	Notes   string  `gorm:"not null; check notes <> ''"`
	Meals   []Meal  `gorm:"many2many:order_meals"`
	Review  *Review `gorm:"foreignKey:OrderID"`
}

type OrderMeal struct {
	OrderID  uint `gorm:"primaryKey"`
	MealID   uint `gorm:"primaryKey"`
	Quantity int  `gorm:"check:quantity >= 1"`
	Status   OrderStatus
}
