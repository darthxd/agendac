package task

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/darthxd/agendac/internal/user"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// recupera o id passado na url da requisição
func parseId(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

// retorna a task pelo id
func GetById(db *gorm.DB, id uint) (*Task, error) {
	var t Task
	if err := db.First(&t, id).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("tarefa com id %d não encontrado", id))
	}
	return &t, nil
}

// cria task para o usuário especificado na requisição usando CreateRequest e a retorna
func createTask(db *gorm.DB, r CreateRequest) (*Task, error) {
	if _, err := user.GetById(db, r.UserID); err != nil {
		return nil, err
	}

	if r.Title == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "titulo da tarefa é obrigatório")
	}

	description := ""
	if r.Description != nil {
		description = *r.Description
	}

	finish_at, err := time.ParseInLocation("2006-01-02", r.FinishAt, time.Local)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "erro ao formatar a data de término da task")
	}

	status := StatusNotStarted
	if r.Status != nil {
		if !validStatus(*r.Status) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "status inválido")
		}
		status = *r.Status
	}

	priority := LowPriority
	if r.Priority != nil {
		if !validPriority(*r.Priority) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "prioridade inválida")
		}
		priority = *r.Priority
	}

	t := &Task{
		UserID:      r.UserID,
		Title:       r.Title,
		Description: &description,
		FinishAt:    finish_at,
		Status:      status,
		Priority:    priority,
	}

	if err := db.Create(t).Error; err != nil {
		return nil, err
	}

	return t, nil
}

// cria task para o usuário salvo no token jwt usando NewRequest e a retorna
func newTask(db *gorm.DB, c *echo.Context, r NewRequest) (*Task, error) {
	me, err := user.ParseUser(db, c)
	if err != nil {
		return nil, err
	}

	if r.Title == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "titulo da tarefa é obrigatório")
	}

	description := ""
	if r.Description != nil {
		description = *r.Description
	}

	finish_at, err := time.ParseInLocation("2006-01-02", r.FinishAt, time.Local)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "erro ao formatar a data de término da task")
	}

	status := StatusNotStarted
	if r.Status != nil {
		if !validStatus(*r.Status) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "status inválido")
		}
		status = *r.Status
	}

	priority := LowPriority
	if r.Priority != nil {
		if !validPriority(*r.Priority) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "prioridade inválida")
		}
		priority = *r.Priority
	}

	t := &Task{
		UserID:      me.ID,
		Title:       r.Title,
		Description: &description,
		FinishAt:    finish_at,
		Status:      status,
		Priority:    priority,
	}

	if err := db.Create(t).Error; err != nil {
		return nil, err
	}

	return t, nil
}
