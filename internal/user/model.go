package user

import "gorm.io/gorm"

const (
	DefaultRole = "default"
	AdminRole   = "admin"
)

func validRole(role string) bool {
	return role == DefaultRole || role == AdminRole
}

func (u User) isAdmin() bool {
	return u.Role == AdminRole
}

// entidade user
type User struct {
	gorm.Model
	DeletedAt gorm.DeletedAt `gorm:"uniqueIndex:idx_email_deleted"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex:idx_email_deleted;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:20;not null;default:default" json:"role"`
}

// dtos
type CreateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User
	Token string `json:"token"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
