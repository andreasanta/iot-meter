package models

import (
	"time"
)

type Account struct {
	ID         string `gorm:"primaryKey"`
	Email      string
	Password   string
	PlanType   string
	TotalUsers uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
