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

// CreateTodo godoc
// @Summary      Create todo
// @Description  Create todo with title and description
// @Tags         todo
// @Accept       json
// @Produce      json
// @Param        todo   body      CreateTodo  true  "Todo"
// @Success      201  {object}  todo.Todo
// @Failure      400  {object}  Error
// @Failure      500  {object}  Error
// @Router       /todos [post]
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

// GetAllTodos godoc
// @Summary      Get all todos
// @Description  Get all todos
// @Tags         todo
// @Accept       json
// @Produce      json
// @Success      200  {array}  todo.Todo
// @Failure      500  {object}  Error
// @Router       /todos [get]
func (handler *TodoHandler) GetAll(ctx *gin.Context) {
	items, err := handler.service.GetAll(ctx)
	if err != nil {
		newError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// GetTodoById godoc
// @Summary      Get todo by ID
// @Description  Get todo by ID
// @Tags         todo
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  todo.Todo
//  @Failure     400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /todos/{id} [get]
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

// UpdateTodo godoc
// @Summary      Update todo
// @Description  Update todo with new title and description
// @Tags         todo
// @Accept       json
// @Produce      json
// @Param        todo   body      UpdateTodo  true  "Todo"
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  todo.Todo
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /todos/{id} [put]
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

// DeleteTodo godoc
// @Summary      Delete todo
// @Description  Delete todo by ID
// @Tags         todo
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      204  {object}  todo.Todo
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /todos/{id} [delete]
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