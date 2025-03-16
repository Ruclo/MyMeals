package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"strings"
)

type Review struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OrderID   uint           `gorm:"unique; not null; constraint: OnDelete:CASCADE, OnUpdate:CASCADE; references:orders(ID)" json:"-"`
	Rating    int            `gorm:"check:rating >= 1 AND rating <= 5" json:"rating" binding:"required"`
	Comment   *string        `json:"comment"`
	PhotoURLs pq.StringArray `gorm:"type:text[]; not null" json:"photo_urls"`
}

func (r *Review) BeforeCreate(db *gorm.DB) error {
	if r.PhotoURLs == nil {
		r.PhotoURLs = pq.StringArray{}
	}

	if r.Comment != nil && len(strings.TrimSpace(*r.Comment)) == 0 {
		r.Comment = nil
	}
	return nil
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
