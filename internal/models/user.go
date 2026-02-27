// internal/models/user.go
package models

import "time"

type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Email       string    `json:"email" gorm:"uniqueIndex;not null"`
	Password    string    `json:"-"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	IsVerified  bool      `json:"isVerified" gorm:"default:false"`
	VerifyCode  string    `json:"-"`
	CodeExpires time.Time `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
