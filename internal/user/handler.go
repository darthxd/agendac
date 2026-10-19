package user

import (
	"fmt"
	"net/http"

	"github.com/darthxd/agendac/internal/auth"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	key string
}

func NewHandler(db *gorm.DB, key string) *Handler {
	return &Handler{db: db, key: key}
}

func (h *Handler) RegisterRoutes(api *echo.Group) {
	users := api.Group("/users")

	users.GET("", h.getAll)
	users.GET("/:id", h.getOne, echojwt.WithConfig(echojwt.Config{SigningKey: []byte(h.key)}))
	users.POST("", h.create)
	users.PUT("/:id", h.update)
	users.DELETE("/:id", h.delete)

	auth := api.Group("/auth")
	auth.POST("/login", h.login)
	auth.POST("/register", h.register)
}

// GET /users -> retorna todos os users
func (h *Handler) getAll(c *echo.Context) error {
	var users []User
	if err := h.db.Find(&users).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}

// GET /users/:id -> retorna o user com o id passado na url
func (h *Handler) getOne(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := getById(h.db, uint(id))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

// POST /users -> cria um novo user e o retorna
func (h *Handler) create(c *echo.Context) error {
	var r CreateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := createUser(h.db, r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, u)
}

// PUT /users/:id -> encontra o user com o id passado na url, atualiza com os dados da requisição e o retorna
func (h *Handler) update(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	var r UpdateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := updateUser(h.db, uint(id), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, u)
}

// DELETE /users/:id -> deleta o user com o id passado na url
func (h *Handler) delete(c *echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := getById(h.db, uint(id))
	if err != nil {
		return err
	}

	if err := h.db.Delete(&u).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("usuário com id %d deletado com sucesso", id)})
}

// POST /auth/login -> faz o login do usuário recebendo email e senha e retornando o usuário e o token jwt
func (h *Handler) login(c *echo.Context) error {
	var r LoginRequest
	if err := c.Bind(&r); err != nil {
		return err
	}

	u, err := getByEmail(h.db, r.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "usuário ou senha incorretos")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(r.Password)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "usuário ou senha incorretos")
	}

	token, err := auth.GenerateToken(u.ID, u.Role)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, LoginResponse{
		User:  *u,
		Token: token,
	})
}

// POST /auth/register -> faz o registro do usuário, mas sempre com role default, e o retorna
func (h *Handler) register(c *echo.Context) error {
	var r RegisterRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := registerUser(h.db, r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, u)
}
