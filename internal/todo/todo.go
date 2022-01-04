package todo

type Todo struct {
	Id          int64  `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
}

func NewTodo(title string, description string) (Todo, error) {
	if len(title) < 1 || len(title) > 255 {
		return Todo{}, newError(ErrValidation, "title should have length between 1 and 255 chars", nil)
	}

	if len(description) < 1 || len(description) > 255 {
		return Todo{}, newError(ErrValidation, "description should have length between 1 and 255 chars", nil)
	}
	return Todo{
		Title:       title,
		Description: description,
	}, nil
}
