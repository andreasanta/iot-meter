package models

import (
	"time"
)

type Account struct {
	ID         string 	`gorm:"primaryKey" gorm:"not null"`
	Email      string
	Password   string
	PlanType   string 	`gorm:"default:Basic" gorm:"not null"`
	TotalUsers uint 	`gorm:"default:0" gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (a *Account) CanAcceptMoreUsers() (bool) {
	if a.PlanType == "Basic" && a.TotalUsers >= 100 {
		return false
	}

	return true
}

func (a *Account) IncreaseUserCount() {
	a.TotalUsers += 1
}

func (a *Account) UpgradePlan() (bool) {

	if a.PlanType != "Basic" {
		return false
	}

	a.PlanType = "Enterprise"

	return true
}


