package main

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET не установлен!")
	}
	jwtKey = []byte(secret)
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// ---------------- JWT ----------------

func GenerateTokens(userID uint) (accessToken, refreshToken string, err error) {
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessToken, err = at.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = rt.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ---------------- User ----------------

func RegisterUser(username, email, password string) (*User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: username,
		Email:    email,
		Password: string(hashed),
	}

	if err := DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func LoginUser(login, password string) (*User, error) {
	var user User
	// WHERE email = ? OR username = ?
	if err := DB.Where("email = ? OR username = ?", login, login).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Проверяем пароль
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

func ResetPassword(email, newPassword string) error {
	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.Password = string(hashed)
	return DB.Save(&user).Error
}
