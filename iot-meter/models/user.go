package models

import (
	"time"
)

type User struct {
	ID string				`gorm:"primaryKey"`
	IotAccountID string
	IotAccount Account		`gorm:"index" gorm:"foreignKey:IotAccountID" gorm:"references:ID"`
	CreatedAt time.Time
  	UpdatedAt time.Time
}