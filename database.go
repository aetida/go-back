package main

import (
	"fmt"

	"github.com/glebarez/sqlite" // чисто-Go драйвер sqlite без CGO :contentReference[oaicite:1]{index=1}
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	// Используем относительный путь "./app.db" — база будет в директории проекта
	DB, err = gorm.Open(sqlite.Open("./app.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// Авто‑миграция моделей, чтобы таблицы создались
	err = DB.AutoMigrate(&User{}, &Session{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}
