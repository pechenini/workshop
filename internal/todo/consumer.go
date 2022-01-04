package todo

type Consumer interface {
	Consume() error
}