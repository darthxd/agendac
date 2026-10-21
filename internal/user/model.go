package user

import (
	"gorm.io/gorm"
)

const (
	DefaultRole = "default"
	AdminRole   = "admin"
)

func ValidRole(r string) bool {
	return r == DefaultRole || r == AdminRole
}

func (u User) IsAdmin() bool {
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
	Name     string `json:"name" example:"João Silva"`
	Email    string `json:"email" example:"joao@email.com"`
	Password string `json:"password" example:"senha123"`
	Role     string `json:"role" enums:"default,admin" example:"default"`
}

type UpdateRequest struct {
	Name     *string `json:"name" example:"João Silva"`
	Email    *string `json:"email" example:"joao@email.com"`
	Password *string `json:"password" example:"novaSenha123"`
	Role     *string `json:"role" enums:"default,admin" example:"default"`
}

type UserResponse struct {
	ID    uint   `json:"id" example:"1"`
	Name  string `json:"name" example:"João Silva"`
	Email string `json:"email" example:"joao@email.com"`
	Role  string `json:"role" enum:"default,admin" example:"default"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"joao@email.com"`
	Password string `json:"password" example:"senha123"`
}

type LoginResponse struct {
	UserResponse
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RegisterRequest struct {
	Name     string `json:"name" example:"João Silva"`
	Email    string `json:"email" example:"joao@email.com"`
	Password string `json:"password" example:"senha123"`
}

// ErrorResponse representa o corpo retornado pelo echo em caso de erro (echo.HTTPError)
type ErrorResponse struct {
	Message string `json:"message" example:"descrição do erro"`
}

// MessageResponse representa uma resposta simples com uma mensagem de sucesso
type MessageResponse struct {
	Message string `json:"message" example:"usuário com id 1 deletado com sucesso"`
}
