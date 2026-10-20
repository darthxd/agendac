package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

// recupera as claims do token jwt
func parseClaims(c *echo.Context) (jwt.MapClaims, error) {
	token, err := echo.ContextGet[*jwt.Token](c, "user")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "erro ao recuperar as claims do token jwt")
	}
	return token.Claims.(jwt.MapClaims), nil
}

// recupera os dados do usuário registrado no token jwt
func parseMe(db *gorm.DB, c *echo.Context) (*User, error) {
	claims, err := parseClaims(c)
	if err != nil {
		return nil, err
	}

	idf, ok := claims["user_id"].(float64)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "erro ao recuperar o user_id do jwt")
	}

	u, err := getById(db, uint(idf))
	if err != nil {
		return nil, err
	}

	return u, nil
}

// verifica status de admin
func verifyAdmin(db *gorm.DB, c *echo.Context) error {
	me, err := parseMe(db, c)
	if err != nil {
		return err
	}
	if !me.isAdmin() {
		return echo.NewHTTPError(http.StatusUnauthorized, "você não tem permissão para acessar esse conteúdo")
	}
	return nil
}

// retorna o user pelo id
func getById(db *gorm.DB, id uint) (*User, error) {
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

// parte lógica para atualizar o user pelo id
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
		if !validRole(*r.Role) {
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

// parte lógica para atualizar o user registrado no token jwt
func updateUserMe(db *gorm.DB, c *echo.Context, r UpdateRequest) (*User, error) {
	me, err := parseMe(db, c)
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
	if me.isAdmin() {
		if r.Role != nil {
			if !validRole(*r.Role) {
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

// parte lógica para registrar o user
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
