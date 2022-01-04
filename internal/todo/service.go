package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Service struct {
	repository Repository
	publisher  Publisher
}

func NewService(repository Repository, publisher Publisher) *Service {
	return &Service{repository: repository, publisher: publisher}
}

func (s *Service) Create(ctx context.Context, title string, description string) (Todo, error) {
	todo, err := NewTodo(title, description)
	if err != nil {
		return Todo{}, err
	}

	todoId, err := s.repository.Create(ctx, todo)
	if err != nil {
		return Todo{}, newError(ErrInternal, "failed to create todo", err)
	}

	todo.Id = todoId

	err = s.publisher.Publish(ctx, newEventCreate(todo))
	if err != nil {
		return todo, newError(ErrEventPublish, "failed to publish event", err)
	}

	return todo, nil
}

func (s *Service) GetAll(ctx context.Context) ([]Todo, error) {
	todos, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, newError(ErrInternal, "failed to get all todos", err)
	}

	return todos, nil
}

func (s *Service) GetById(ctx context.Context, id int64) (Todo, error) {
	todo, err := s.repository.GetById(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, newError(ErrNotFound, fmt.Sprintf("todo with id %d is not found", id), err)
	}

	if err != nil {
		return Todo{}, newError(ErrInternal, "failed to get todo by id", err)
	}

	return todo, nil
}

func (s *Service) Update(ctx context.Context, todo Todo) error {
	err := s.Update(ctx, todo)
	if err != nil {
		return newError(ErrInternal, "failed to update todo", err)
	}

	err = s.publisher.Publish(ctx, newEventUpdate(todo))
	if err != nil {
		return newError(ErrEventPublish, "failed to publish event", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, todo Todo) error {
	err := s.repository.Delete(ctx, todo.Id)
	if err != nil {
		return newError(ErrInternal, "failed to delete todo", err)
	}

	err = s.publisher.Publish(ctx, newEventDelete(todo))
	if err != nil {
		return newError(ErrEventPublish, "failed to publish event", err)
	}

	return nil
}
