package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

// Role represents the two types of authenticated users.
type Role string

const (
	AdminRole        Role = "AdminRole"
	RegularStaffRole Role = "Regular Staff"
)

// Valid checks if the Role value is one of the predefined valid roles and returns an error if it is invalid.
func (m Role) Valid() error {
	switch m {
	case AdminRole, RegularStaffRole:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid role %s", m))
	}
}

// Scan assigns a value to the Role type, converting from string or []byte, and validates it using the Valid method.
func (m *Role) Scan(value interface{}) error {
	if value == nil {
		*m = ""
		return nil
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid scan source for Role")
		}
		str = string(bytes)
	}

	*m = Role(str)
	return m.Valid()
}

// Value converts the Role type to a driver.Value for database storage, returning an error if the role is invalid.
func (m Role) Value() (driver.Value, error) {
	if err := m.Valid(); err != nil {
		return nil, err
	}
	return string(m), nil
}

type User struct {
	Username string `gorm:"primaryKey"`
	Password string `gorm:"not null"`
	Role     Role   `gorm:"not null; default: 'Regular Staff'"`
}
