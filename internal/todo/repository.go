package todo

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, todo Todo) (int64, error)
	GetAll(ctx context.Context) ([]Todo, error)
	GetById(ctx context.Context, id int64) (Todo, error)
	Update(ctx context.Context, todo Todo) error
	Delete(ctx context.Context, id int64) error
}
