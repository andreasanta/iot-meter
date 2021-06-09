package models

import (
	"time"
)

type User struct {
	ID           string `gorm:"primaryKey"`
	IotAccountID string
	IotAccount   Account `gorm:"foreignKey:IotAccountID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
