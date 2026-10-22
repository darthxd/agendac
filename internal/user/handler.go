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
	cfg := echojwt.Config{SigningKey: []byte(h.key)}

	users := api.Group("/users", echojwt.WithConfig(cfg))

	users.GET("", h.getAll)
	users.GET("/:id", h.getOne)
	users.GET("/me", h.getMe)
	users.POST("", h.create)
	users.PATCH("/:id", h.update)
	users.PATCH("/me", h.updateMe)
	users.DELETE("/:id", h.delete)

	auth := api.Group("/auth")
	auth.POST("/login", h.login)
	auth.POST("/register", h.register)
}

// getAllUser godoc
//
//	@Summary		Listar usuários
//	@Description	Retorna uma lista com todos os usuários cadastrados
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		UserResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/users [get]
func (h *Handler) getAll(c *echo.Context) error {
	var users []User
	if err := h.db.Find(&users).Error; err != nil {
		return err
	}

	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = bindResponse(&u)
	}

	return c.JSON(http.StatusOK, res)
}

// getOneUser godoc
//
//	@Summary		Buscar usuário por id
//	@Description	Retorna o usuário com o id passado na url. Requer permissão de admin
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID do usuário"
//	@Success		200	{object}	UserResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/users/{id} [get]
func (h *Handler) getOne(c *echo.Context) error {
	me, err := ParseUser(h.db, c)
	if err != nil {
		return err
	}

	if err := me.VerifyAdmin(); err != nil {
		return err
	}

	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := GetById(h.db, uint(id))
	if err != nil {
		return err
	}

	res := bindResponse(u)
	return c.JSON(http.StatusOK, res)
}

// getMeUser godoc
//
//	@Summary		Buscar usuário autenticado
//	@Description	Retorna o usuário salvo no jwt da requisição
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/users/me [get]
func (h *Handler) getMe(c *echo.Context) error {
	u, err := ParseUser(h.db, c)
	if err != nil {
		return err
	}

	res := bindResponse(u)
	return c.JSON(http.StatusOK, res)
}

// createUser godoc
//
//	@Summary		Criar usuário
//	@Description	Cria um novo usuário e o retorna. Requer permissão de admin
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRequest	true	"Dados do usuário"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/users [post]
func (h *Handler) create(c *echo.Context) error {
	me, err := ParseUser(h.db, c)
	if err != nil {
		return err
	}

	if err := me.VerifyAdmin(); err != nil {
		return err
	}

	var r CreateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := createUser(h.db, r)
	if err != nil {
		return err
	}

	res := bindResponse(u)

	return c.JSON(http.StatusCreated, res)
}

// updateUser godoc
//
//	@Summary		Atualizar usuário
//	@Description	Atualiza o usuário com o id passado na url e o retorna. Requer permissão de admin
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"ID do usuário"
//	@Param			request	body		UpdateRequest	true	"Campos a serem atualizados"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/users/{id} [patch]
func (h *Handler) update(c *echo.Context) error {
	me, err := ParseUser(h.db, c)
	if err != nil {
		return err
	}

	if err := me.VerifyAdmin(); err != nil {
		return err
	}

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

	res := bindResponse(u)
	return c.JSON(http.StatusOK, res)
}

// updateMeUser godoc
//
//	@Summary		Atualizar usuário autenticado
//	@Description	Atualiza o usuário salvo no jwt da requisição. Um usuário só pode alterar seu role caso ele mesmo seja admin
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpdateRequest	true	"Campos a serem atualizados"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/users/me [patch]
func (h *Handler) updateMe(c *echo.Context) error {
	var r UpdateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	me, err := updateUserMe(h.db, c, r)
	if err != nil {
		return err
	}

	res := bindResponse(me)
	return c.JSON(http.StatusOK, res)
}

// deleteUser godoc
//
//	@Summary		Deletar usuário
//	@Description	Deleta o usuário com o id passado na url. Requer permissão de admin
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID do usuário"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/users/{id} [delete]
func (h *Handler) delete(c *echo.Context) error {
	me, err := ParseUser(h.db, c)
	if err != nil {
		return err
	}

	if err := me.VerifyAdmin(); err != nil {
		return err
	}

	id, err := parseId(c)
	if err != nil {
		return err
	}

	u, err := GetById(h.db, uint(id))
	if err != nil {
		return err
	}

	if err := h.db.Delete(&u).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("usuário com id %d deletado com sucesso", id)})
}

// login godoc
//
//	@Summary		Login
//	@Description	Faz o login do usuário recebendo email e senha e retorna o usuário e o token jwt
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Credenciais"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/auth/login [post]
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

	token, err := auth.GenerateToken(u.ID)
	if err != nil {
		return err
	}

	res := bindResponse(u)
	return c.JSON(http.StatusOK, LoginResponse{
		UserResponse: res,
		Token:        token,
	})
}

// register godoc
//
//	@Summary		Registrar usuário
//	@Description	Faz o registro do usuário, sempre com role default, e o retorna
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Dados do usuário"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/auth/register [post]
func (h *Handler) register(c *echo.Context) error {
	var r RegisterRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	u, err := registerUser(h.db, r)
	if err != nil {
		return err
	}

	res := bindResponse(u)
	return c.JSON(http.StatusCreated, res)
}
