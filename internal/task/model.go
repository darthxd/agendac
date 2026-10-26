package task

import (
	"time"

	"github.com/darthxd/agendac/internal/user"
	"gorm.io/gorm"
)

const (
	StatusNotStarted = "not_started"
	StatusOngoing    = "ongoing"
	StatusFinished   = "finished"

	HighPriority   = "high"
	MediumPriority = "medium"
	LowPriority    = "low"
)

func validStatus(s string) bool {
	return s == StatusNotStarted || s == StatusOngoing || s == StatusFinished
}

func validPriority(p string) bool {
	return p == HighPriority || p == MediumPriority || p == LowPriority
}

type Task struct {
	gorm.Model
	UserID      uint      `json:"user_id" gorm:"index"`
	User        user.User `json:"user" gorm:"foreignKey:UserID"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	FinishAt    time.Time `json:"finish_at"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Color       string    `json:"color"`
}

// dtos
type CreateRequest struct {
	UserID      uint    `json:"user_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	FinishAt    string  `json:"finish_at"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	Color       string  `json:"color"`
}

type NewRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	FinishAt    string  `json:"finish_at"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	Color       string  `json:"color"`
}

type UpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	FinishAt    *string `json:"finish_at"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	Color       *string `json:"color"`
}

type TaskResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	FinishAt    time.Time `json:"finish_at"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Color       string    `json:"color"`
}

type TaskFilter struct {
	Status     string `query:"status"`
	Priority   string `query:"priority"`
	UserID     uint   `query:"user_id"`
	SearchTerm string `query:"q"`
}

// ErrorResponse representa o corpo retornado pelo echo em caso de erro (echo.HTTPError)
type ErrorResponse struct {
	Message string `json:"message" example:"descrição do erro"`
}

// MessageResponse representa uma resposta simples com uma mensagem de sucesso
type MessageResponse struct {
	Message string `json:"message" example:"usuário com id 1 deletado com sucesso"`
}
