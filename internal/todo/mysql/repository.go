package mysql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"workshop/internal/todo"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, todo todo.Todo) (int64, error) {
	result, err := r.db.NamedExecContext(ctx, "INSERT INTO todos(title, description) VALUES (:title, :description)", todo)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetAll(ctx context.Context) ([]todo.Todo, error) {
	todos := make([]todo.Todo, 0)
	err := r.db.SelectContext(ctx, &todos, "SELECT * from todos")
	return todos, err
}

func (r *Repository) GetById(ctx context.Context, id int64) (todo.Todo, error) {
	var todoItem todo.Todo
	err := r.db.GetContext(ctx, &todoItem, "SELECT * from todos WHERE id=?", id)
	return todoItem, err
}

func (r *Repository) Update(ctx context.Context, todo todo.Todo) error {
	_, err := r.db.NamedExecContext(ctx, "UPDATE todos SET title=:title, description=:description WHERE id=:id", todo)
	return err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM todos WHERE id=?", id)
	return err
}