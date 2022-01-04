package http

import (
	"context"
	"workshop/internal/todo"
)

type TodoService interface {
	Create(ctx context.Context, title string, description string) (todo.Todo, error)
	GetAll(ctx context.Context) ([]todo.Todo, error)
	GetById(ctx context.Context, id int64) (todo.Todo, error)
	Update(ctx context.Context, todo todo.Todo) error
	Delete(ctx context.Context, todo todo.Todo) error
}
