package task

import (
	"fmt"
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
	tasks.PATCH("/:id", h.update)
	tasks.DELETE("/:id", h.delete)

	tasks.POST("/new", h.new)
}

// getAllTask godoc
//
//	@Summary		Listar tarefas
//	@Description	Retorna uma lista com todas as tarefas cadastradas
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query		string	false	"status da tarefa"
//	@Param			priority	query		string	false	"priority da tarefa"
//	@Param			user_id		query		int		false	"id do usuário"
//	@Param			q			query		string	false	"termo para pesquisa (titulo e descrição)"
//	@Success		200			{array}		TaskResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Router			/tasks [get]
func (h *Handler) getAll(c *echo.Context) error {
	var f TaskFilter

	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "erro ao processar os filtros da pesquisa")
	}

	query := f.ApplyFilters(h.db)

	var tasks []Task
	if err := query.Preload("User").Find(&tasks).Error; err != nil {
		return err
	}

	res := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		res[i] = bindResponse(&t)
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

	res := bindResponse(t)
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

	res := bindResponse(t)
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

	res := bindResponse(t)
	return c.JSON(http.StatusCreated, res)
}

// updateTask godoc
//
//	@Summary		Atualiza tarefa
//	@Description	Atualiza uma tarefa. Caso o usuário não seja admin, ele só pode alterar suas próprias tarefas
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"ID da tarefa"
//	@Param			request	body		UpdateRequest	true	"Dados da tarefa"
//	@Success		200		{object}	TaskResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/tasks/{id} [patch]
func (h *Handler) update(c *echo.Context) error {
	me, err := user.ParseUser(h.db, c)
	if err != nil {
		return err
	}

	tid, err := parseId(c)
	if err != nil {
		return err
	}

	tprev, err := GetById(h.db, tid)
	if err != nil {
		return err
	}

	if me.ID != tprev.UserID {
		if err := me.VerifyAdmin(); err != nil {
			return err
		}
	}

	var r UpdateRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "corpo da requisição inválido")
	}

	t, err := updateTask(h.db, tid, r)
	if err != nil {
		return err
	}

	res := bindResponse(t)
	return c.JSON(http.StatusOK, res)
}

// deleteTask godoc
//
//	@Summary		Deletar tarefa
//	@Description	Deleta a tarefa com o id passado na url. Caso o usuário não seja admin, ele só pode deletar suas próprias tarefas
//	@Tags			tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"ID da tarefa"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/tasks/{id} [delete]
func (h *Handler) delete(c *echo.Context) error {
	me, err := user.ParseUser(h.db, c)
	if err != nil {
		return err
	}

	tid, err := parseId(c)
	if err != nil {
		return err
	}

	t, err := GetById(h.db, tid)
	if err != nil {
		return err
	}

	if me.ID != t.UserID {
		if err := me.VerifyAdmin(); err != nil {
			return err
		}
	}

	if err := h.db.Delete(&t).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("tarefa com id %d deletado com sucesso", tid)})
}
