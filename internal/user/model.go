package user

import "gorm.io/gorm"

const (
	DefaultRole = "default"
	AdminRole   = "admin"
)

type User struct {
	gorm.Model
	Name     string `gorm:"size:100;not null" json:"name"`
	Email    string `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	Role     string `gorm:"size:20;not null;default:default" json:"role"`
}
