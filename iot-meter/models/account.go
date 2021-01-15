package models


import (
	"time"
)

type Account struct {
	Email string				`gorm:"primaryKey"`
	ID string					`gorm:"primaryKey"`
	Password string
	PlanType string
	TotalUsers uint
	CreatedAt time.Time
  	UpdatedAt time.Time
}