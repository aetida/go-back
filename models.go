package main

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Email    string `gorm:"uniqueIndex"`
	Password string
}

type Session struct {
	gorm.Model
	SessionID       string `gorm:"uniqueIndex"`
	ClientConnected bool
	PhoneConnected  bool
	CreatedAt       time.Time
}
