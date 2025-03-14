package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

type Review struct {
	ID        uint     `gorm:"primaryKey"`
	OrderID   uint     `gorm:"unique; not null; constraint: OnDelete:CASCADE, OnUpdate:CASCADE; references:orders(ID)"`
	Rating    int      `gorm:"check:rating >= 1 AND rating <= 5"`
	Comment   string   `gorm:"not null; check:comment <> ''"`
	PhotoURLs []string `gorm:"type:text[]; not null"`
}

type PhotoURLs []string

func (p PhotoURLs) Valid() error {
	for _, url := range p {
		if strings.TrimSpace(url) == "" {
			return errors.New("Invalid photo URL")
		}
	}
	return nil
}

func (p *PhotoURLs) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal PhotoURLs value")
	}

	return json.Unmarshal(bytes, p)

}

func (p PhotoURLs) Value() (driver.Value, error) {
	if err := p.Valid(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}
