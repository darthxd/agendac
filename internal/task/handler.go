package task

import (
	"net/http"

	"github.com/darthxd/agendac/internal/user"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
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

	tasks := api.Group("/tasks", echojwt.WithConfig(cfg))
	tasks.GET("", h.getAll)
	tasks.GET("/:id", h.getOne)
	tasks.POST("", h.create)

	tasks.POST("/new", h.new)
}

// getAllTask godoc
//
//	@Summary		Listar tarefas
//	@Description	Retorna uma lista com todas as tarefas cadastradas
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		TaskResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/tasks [get]
func (h *Handler) getAll(c *echo.Context) error {
	var tasks []Task
	if err := h.db.Find(&tasks).Error; err != nil {
		return err
	}

	res := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		res[i] = TaskResponse{
			ID:          t.ID,
			UserID:      t.UserID,
			Title:       t.Title,
			Description: *t.Description,
			CreatedAt:   t.CreatedAt,
			FinishAt:    t.FinishAt,
			Status:      t.Status,
			Priority:    t.Priority,
		}
	}

	return c.JSON(http.StatusOK, res)
}

// getOneTask godoc
//
//	@Summary		Buscar tarefa por id
//	@Description	Retorna a tarefa com o id passado na url. Requer permissão de admin
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID da tarefa"
//	@Success		200	{object}	TaskResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/tasks/{id} [get]
func (h *Handler) getOne(c *echo.Context) error {
	me, err := user.ParseUser(h.db, c)
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

	t, err := GetById(h.db, id)
	if err != nil {
		return err
	}

	res := TaskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: *t.Description,
		CreatedAt:   t.CreatedAt,
		FinishAt:    t.FinishAt,
		Status:      t.Status,
		Priority:    t.Priority,
	}

	return c.JSON(http.StatusOK, res)
}

// createTask godoc
//
//	@Summary		Criar tarefa
//	@Description	Cria uma nova tarefa. Requer admin
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRequest	true	"Dados da tarefa"
//	@Success		200		{object}	TaskResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/tasks [post]
func (h *Handler) create(c *echo.Context) error {
	me, err := user.ParseUser(h.db, c)
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

	t, err := createTask(h.db, r)
	if err != nil {
		return err
	}

	res := TaskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: *t.Description,
		CreatedAt:   t.CreatedAt,
		FinishAt:    t.FinishAt,
		Status:      t.Status,
		Priority:    t.Priority,
	}

	return c.JSON(http.StatusCreated, res)
}

// newTask godoc
//
//	@Summary		Criar tarefa para usuário autenticado
//	@Description	Cria uma nova tarefa relacionada ao usuário salvo no token jwt
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		NewRequest	true	"Dados da tarefa"
//	@Success		200		{object}	TaskResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/tasks/new [post]
func (h *Handler) new(c *echo.Context) error {
	var r NewRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	t, err := newTask(h.db, c, r)
	if err != nil {
		return err
	}

	res := TaskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: *t.Description,
		CreatedAt:   t.CreatedAt,
		FinishAt:    t.FinishAt,
		Status:      t.Status,
		Priority:    t.Priority,
	}

	return c.JSON(http.StatusCreated, res)
}
