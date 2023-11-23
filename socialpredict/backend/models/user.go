package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uint    `json:"id" gorm:"primary_key"`
	Username       string  `json:"username" gorm:"unique;not null"`
	Email          string  `json:"email" gorm:"unique;not null"`
	Password       string  `json:"password,omitempty" gorm:"not null"`
	UserType       string  `json:"usertype" gorm:"unique;not null"`
	AccountBalance float64 `json:"accountBalance"`
}

// HashPassword hashes given password
func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	u.Password = string(bytes)
	return err
}

// CheckPasswordHash checks if provided password is correct
func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
