package todo

import "context"

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}