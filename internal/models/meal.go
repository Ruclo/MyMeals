package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type MealCategory string

// DB Initialisation
const (
	Drinks      MealCategory = "Drinks"
	Starters    MealCategory = "Starters"
	MainCourses MealCategory = "Main Courses"
	SideDishes  MealCategory = "Side Dishes"
	Desserts    MealCategory = "Desserts"
)

func (c MealCategory) Valid() error {
	switch c {
	case Drinks, Starters, MainCourses, SideDishes, Desserts:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid meal category %s", c))
	}
}

func (c *MealCategory) Scan(value interface{}) error {
	if value == nil {
		*c = ""
		return nil
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid scan source for MealCategory")
		}
		str = string(bytes)
	}

	*c = MealCategory(str)
	return c.Valid()
}

func (c MealCategory) Value() (driver.Value, error) {
	if err := c.Valid(); err != nil {
		return nil, err
	}
	return string(c), nil
}

type Meal struct {
	ID          uint            `gorm:"primaryKey;autoIncrement"`
	Name        string          `gorm:"not null; check name <> ''"`
	Category    MealCategory    `gorm:"not null"`
	Description string          `gorm:"not null; check description <> ''"`
	PhotoURL    string          `gorm:"not null; check description <> ''"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2); check price > 0"`
}
