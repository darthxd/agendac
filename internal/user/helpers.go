package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func parseId(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

func (h *Handler) getById(id uint) (*User, error) {
	var user User
	if err := h.db.First(&user, id).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("usuário com id %d não encontrado", id))
	}
	return &user, nil
}

func (h *Handler) create(r CreateRequest) (*User, error) {
	if r.Name == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "nome é obrigatório")
	}
	if r.Email == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "email é obrigatório")
	}
	if r.Password == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "senha é obrigatória")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		Name:     r.Name,
		Email:    r.Email,
		Password: string(hash),
		Role:     DefaultRole,
	}

	if err := h.db.Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
		return nil, err
	}

	return u, nil
}

func (h *Handler) update(id uint, r UpdateRequest) (*User, error) {
	u, err := h.getById(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "nome inválido")
		}
		updates["name"] = name
	}
	if r.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*r.Email))
		if email == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "email inválido")
		}
		updates["email"] = email
	}
	if r.Password != nil {
		name := strings.TrimSpace(*r.Password)
		if name == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "senha inválida")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*r.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password"] = hash
	}
	if r.Role != nil {
		if *r.Role != DefaultRole && *r.Role != AdminRole {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "role inválido")
		}
		updates["role"] = r.Role
	}

	if err := h.db.Model(&u).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
	}

	return u, nil
}
