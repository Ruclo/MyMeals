package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

type Role string

const (
	AdminRole        Role = "AdminRole"
	RegularStaffRole Role = "Regular Staff"
)

func (m Role) Valid() error {
	switch m {
	case AdminRole, RegularStaffRole:
		return nil
	default:
		return errors.New(fmt.Sprintf("Invalid role %s", m))
	}
}

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

func (m Role) Value() (driver.Value, error) {
	if err := m.Valid(); err != nil {
		return nil, err
	}
	return string(m), nil
}

type StaffMember struct {
	Username string `gorm:"primaryKey" json:"username" binding:"required"`
	Password string `gorm:"not null" json:"password" binding:"required,min=8"`
	Role     Role   `gorm:"not null; default: 'Regular Staff'" json:"-"`
}
