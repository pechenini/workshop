package http

type CreateTodo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}