package model

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"size:120;not null;uniqueIndex"`
	Username  string `gorm:"size:60;not null;uniqueIndex"`
	Password  string `gorm:"size:255;not null"` //bcrypt hash
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
