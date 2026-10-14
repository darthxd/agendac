package user

import (
	"fmt"
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

	users.GET("", h.userList)
	users.GET("/:id", h.userGetById)
	users.POST("", h.userCreate)
	users.PUT("/:id", h.userUpdate)
	users.DELETE("/:id", h.userDelete)
}

func (h *Handler) userList(c *echo.Context) error {
	var users []User
	if err := h.db.Find(&users).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}

func (h *Handler) userGetById(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := h.getById(uint(id))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

func (h *Handler) userCreate(c *echo.Context) error {
	var r CreateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := h.create(r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) userUpdate(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	var r UpdateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := h.update(uint(id), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

func (h *Handler) userDelete(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := h.getById(uint(id))
	if err != nil {
		return err
	}

	if err := h.db.Delete(&u).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("usuário com id %d deletado com sucesso", id)})
}
