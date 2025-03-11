package models

type StaffMember struct {
		Username string `gorm:"primaryKey"`
		Password string		//min length 8
}