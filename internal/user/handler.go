package user

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(api *echo.Group) {
	users := api.Group("/users")
	users.GET("", h.List)
}

func (h *Handler) List(c *echo.Context) error {
	var users []User
	if err := h.db.Find(&users).Error; err != nil {
		return err
	}
	return c.JSON(http.StatusOK, users)
}
