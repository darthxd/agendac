package user

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/darthxd/agendac/internal/auth"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// gera o usuário padrão admin caso não exista nenhum usuário no sistema
func SeedAdmin(db *gorm.DB) error {
	var c int64
	if err := db.Model(&User{}).Count(&c).Error; err != nil {
		return err
	}
	if c > 0 {
		return nil
	}

	password, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := User{
		Name:     "admin",
		Email:    "admin@admin.com",
		Password: string(password),
		Role:     AdminRole,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	slog.Info("usuário admin criado com sucesso")
	return nil
}

// recupera o id passado na url da requisição
func parseId(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

// recupera os dados do usuário salvo no token jwt
func ParseUser(db *gorm.DB, c *echo.Context) (*User, error) {
	claims, err := auth.ParseClaims(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "erro ao recuperar as claims do token jwt")
	}

	idf, ok := claims["user_id"].(float64)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "erro ao recuperar o user_id do jwt")
	}

	u, err := GetById(db, uint(idf))
	if err != nil {
		return nil, err
	}

	return u, nil
}

// verifica role do usuário e retorna erro caso não seja admin
func (u User) VerifyAdmin() error {
	if !u.IsAdmin() {
		return echo.NewHTTPError(http.StatusUnauthorized, "você não tem permissão para acessar esse conteúdo")
	}
	return nil
}

// retorna o user pelo id
func GetById(db *gorm.DB, id uint) (*User, error) {
	var u User
	if err := db.First(&u, id).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("usuário com id %d não encontrado", id))
	}
	return &u, nil
}

// retorna o user pelo email
func getByEmail(db *gorm.DB, email string) (*User, error) {
	var u User
	if err := db.Where(&User{Email: email}).First(&u).Error; err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("usuário com email %s não encontrado", email))
	}
	return &u, nil
}

// cria o user usando CreateRequest e o retorna
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
		if !ValidRole(r.Role) {
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

// atualiza o user recebendo um id e UpdateRequest e o retorna
func updateUser(db *gorm.DB, id uint, r UpdateRequest) (*User, error) {
	u, err := GetById(db, id)
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
		if !ValidRole(*r.Role) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "role inválido")
		}
		updates["role"] = r.Role
	}

	if err := db.Model(&u).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("erro desconhecido: %s", err))
	}

	return u, nil
}

// atualiza o user salvo no token jwt usando UpdateRequest e o retorna
func updateUserMe(db *gorm.DB, c *echo.Context, r UpdateRequest) (*User, error) {
	me, err := ParseUser(db, c)
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
	if me.IsAdmin() {
		if r.Role != nil {
			if !ValidRole(*r.Role) {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "role inválido")
			}
			updates["role"] = r.Role
		}
	}

	if err := db.Model(&me).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, echo.NewHTTPError(http.StatusConflict, "email já cadastrado")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("erro desconhecido: %s", err))
	}

	return me, nil
}

// registra o user usando RegisterRequest e o retorna
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
