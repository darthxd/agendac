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

// recupera o id passado na url da requisição
func parseId(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

// retorna o user pelo id
func getById(db *gorm.DB, id uint) (*User, error) {
	var u User
	if err := db.First(&u, id).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("usuário com id %d não encontrado", id))
	}
	return &u, nil
}

func getByEmail(db *gorm.DB, email string) (*User, error) {
	var u User
	if err := db.Where(&User{Email: email}).First(&u).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("usuário com email %s não encontrado", email))
	}
	return &u, nil
}

// parte lógica para criar o user
func createUser(db *gorm.DB, r CreateRequest) (*User, error) {
	if r.Name == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "nome é obrigatório")
	}
	if r.Email == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "email é obrigatório")
	}
	if r.Password == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "senha é obrigatória")
	}
	role := DefaultRole
	if r.Role != "" {
		if !validRole(r.Role) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "role inválido")
		}
		role = r.Role
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		Name:     r.Name,
		Email:    r.Email,
		Password: string(hash),
		Role:     role,
	}

	if err := db.Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
		return nil, err
	}

	return u, nil
}

// parte lógica para atualizar o user
func updateUser(db *gorm.DB, id uint, r UpdateRequest) (*User, error) {
	u, err := getById(db, id)
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

	if err := db.Model(&u).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
	}

	return u, nil
}

func registerUser(db *gorm.DB, r RegisterRequest) (*User, error) {
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

	if err := db.Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
		return nil, err
	}

	return u, nil
}
