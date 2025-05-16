package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MealCategory represents one of the 5 categories of meals.
type MealCategory string

// DB Initialisation
const (
	Drinks      MealCategory = "Drinks"
	Starters    MealCategory = "Starters"
	MainCourses MealCategory = "Main Courses"
	SideDishes  MealCategory = "Side Dishes"
	Desserts    MealCategory = "Desserts"
)

// Valid checks if the MealCategory is one of the predefined valid categories, returning an error if invalid.
func (c MealCategory) Valid() error {
	switch c {
	case Drinks, Starters, MainCourses, SideDishes, Desserts:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid meal category %s", c))
	}
}

// Scan implements the sql.Scanner interface, allowing MealCategory to be scanned from database values.
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

// Value converts the MealCategory to a driver.Value for database storage, returning an error if the value is invalid.
func (c MealCategory) Value() (driver.Value, error) {
	if err := c.Valid(); err != nil {
		return nil, err
	}
	return string(c), nil
}

type Meal struct {
	ID          uint            `gorm:"primaryKey;autoIncrement"`
	Name        string          `gorm:"not null; check: name <> ''"`
	Category    MealCategory    `gorm:"not null" json:"category"`
	Description string          `gorm:"not null; check: description <> ''"`
	ImageURL    string          `gorm:"not null; check: image_url <> ''"`
	Price       decimal.Decimal `gorm:"type:numeric(10,2); check: price > 0"`
	DeletedAt   gorm.DeletedAt  `json:"-"`
}
