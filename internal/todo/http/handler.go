package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"workshop/internal/todo"
)

type TodoService interface {
	Create(ctx context.Context, title string, description string) (todo.Todo, error)
	GetAll(ctx context.Context) ([]todo.Todo, error)
	GetById(ctx context.Context, id int64) (todo.Todo, error)
	Update(ctx context.Context, todo todo.Todo) error
	Delete(ctx context.Context, todo todo.Todo) error
}

type TodoHandler struct {
	service TodoService
}

func NewTodoHandler(service TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (handler *TodoHandler) Create(ctx *gin.Context) {
	var createRequest CreateTodo
	if err := ctx.ShouldBindJSON(&createRequest); err != nil {
		badRequest(ctx, err)
		return
	}

	item, err := handler.service.Create(ctx, createRequest.Title, createRequest.Description)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (handler *TodoHandler) GetAll(ctx *gin.Context) {
	items, err := handler.service.GetAll(ctx)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, items)
}

func (handler *TodoHandler) GetById(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		badRequest(ctx, err)
		return
	}

	item, err := handler.service.GetById(ctx, id)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (handler *TodoHandler) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		badRequest(ctx, err)
		return
	}

	var updateRequest UpdateTodo
	if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
		badRequest(ctx, err)
		return
	}

	existingTodo, err := handler.service.GetById(ctx, id)
	if err != nil {
		newError(ctx, err)
		return
	}

	updatedTodo, err := todo.NewTodo(updateRequest.Title, updateRequest.Description)
	if err != nil {
		newError(ctx, err)
		return
	}

	updatedTodo.Id = existingTodo.Id

	err = handler.service.Update(ctx, updatedTodo)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, updatedTodo)
}

func (handler *TodoHandler) Delete(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		badRequest(ctx, err)
		return
	}

	item, err := handler.service.GetById(ctx, id)
	if err != nil {
		newError(ctx, err)
		return
	}

	err = handler.service.Delete(ctx, item)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, item)
}