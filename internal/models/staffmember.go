package models

type StaffMember struct {
	Username string `gorm:"primaryKey"`
	Password string `gorm:"not null"`
}
