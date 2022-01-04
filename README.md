# Workshop

# Инициализация проекта

Создаем папку с названием `workshop` и заходим в нее.

```bash
mkdir workshop && cd workshop
```

Теперь нужно инициализировать модуль. Делается это с помощью команды:

```bash
go mod init workshop
```

Внутри папки должны будем увидеть файл `go.mod` с контентом:

```
module workshop

go 1.17

```

Это аналог **composer.json.**
После добавления зависимостей в наш проект увидим еще один файл `go.sum` (аналог composer.lock).

# Структура проекта

Будем использовать [https://github.com/golang-standards/project-layout](https://github.com/golang-standards/project-layout) .

Набросаем примерную структуру проекта:

```
|____cmd
| |____server
| |____consumer
|____migrations
|____go.mod
|____docs
|____internal
| |____todo
| | |____http
| | |____kafka
| | |____mysql
```

Пройдемся более детально по каждой папке.

- `cmd/server` -> точка входа для запуска хттп сервера.
- `cmd/consumer` -> точка входа для запуска консьюмера сообщений из кафки.
- `migrations` -> место хранения миграций всего проекта.
- `docs` -> место хранения документации(swagger).
- `internal/todo/` -> модуль с логикой todo.
- `internal/todo/http` -> http адаптер к логике todo. Взаимодействие бизнес-логики todo через http интерфейс будет проиcходить тут.
- `internal/todo/kafka` -> пакет c реализаций паблишера ивентов в кафку.
- `internal/todo/mysql` -> пакет c реализаций репозитория для mysql.

# Инфраструктура

В корне проекта есть файл `docker-compose.yml`. В нем находятся контейнеры MySQL, zookeeper, kafka.

Поднимите контейнеры:

```bash
docker-compose up -d
```

Создадим топик в кафке:

```bash
docker exec -it workshop_kafka_1 /bin/sh
/opt/bitnami/kafka/bin/kafka-topics.sh --create --partitions 1 --replication-factor 1 --topic todo --bootstrap-server localhost:9092
```

Создадим файл `.env` из `.env.example`:

```bash
cp .env.example .env
```

Чтение из топика:

```bash
docker exec -it workshop_kafka_1 /bin/sh
/opt/bitnami/kafka/bin/kafka-console-consumer.sh --topic todo --from-beginning --bootstrap-server localhost:9092
```

# Todo error wrapper

Опишем вероятные ошибки и создадим свой враппер для ошибок в файле `internal/todo/errors.go`:

```go
package todo

import (
	"fmt"
)

const (
	ErrValidation   = iota + 1 //starting from 1 assign to consts below val++
	ErrInternal                // 2
	ErrEventPublish            // 3
	ErrNotFound                // 4
)

type Error struct {
	Code int
	Msg  string
	Err  error
}

func newError(code int, msg string, err error) error {
	return &Error{
		Code: code,
		Msg:  msg,
		Err:  err,
	}
}

func (err *Error) Unwrap() error {
	return err.Err
}

func (err *Error) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("%s: %s", err.Msg, err.Err.Error())
	}
	return err.Msg
}
```

# Модели Todo и Event

Внутри пакета `internal/todo` создадим файл `todo.go` и определим нашу модель **Todo**:

```go
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
```

Там же создадим еще один файл `event.go` и определим структуру для событий, которые мы будем отправлять в кафку:

```go
package todo

const eventCreate = "create"
const eventUpdate = "update"
const eventDelete = "delete"

type Event struct {
	Event string `json:"event"`
	Todo  Todo   `json:"todo"`
}

func newEvent(event string, todo Todo) Event {
	return Event{Event: event, Todo: todo}
}

func newEventCreate(todo Todo) Event {
	return newEvent(eventCreate, todo)
}

func newEventUpdate(todo Todo) Event {
	return newEvent(eventUpdate, todo)
}

func newEventDelete(todo Todo) Event {
	return newEvent(eventDelete, todo)
}
```

# Определим интерфейсы Todo модуля

Создадим 3 файла в рамках пакета `internal/todo`:

- `consumer.go`
- `publisher.go`
- `repository.go`

Внутри `consumer.go` определим интерфейс для консьюмера:

```go
package todo

type Consumer interface {
	Consume() error
}
```

Реализация будет лежать в `internal/todo/kafka`.

Внутри `publisher.go` определим интерфейс для паблишера:

```go
package todo

import "context"

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}
```

Реализация будет лежать в `internal/todo/kafka`.

Внутри `repository.go` определим интерфейс для репозитория:

```go
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
```

Реализация будет лежать в `internal/todo/mysql`.

# MySQL реализация Repository

Установим зависимости

```
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
```

Создадим файл `internal/todo/mysql/repository.go` :

```go
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
```

## Миграции

Создадим 2 файла внутри папки `migrations`: `1_create_todo_table.up.sql` и `1_create_todo_table.down.sql` .

```sql
CREATE TABLE IF NOT EXISTS todos
(
    id          serial NOT NULL PRIMARY KEY,
    title       varchar(255),
    description text(255)
);
```

```sql
DROP TABLE IF EXISTS todos;
```

# Kafka реализация Publisher

Установим зависимости

```
go get github.com/segmentio/kafka-go
```

Создадим файл `internal/todo/kafka/publisher.go` :

```go
package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"strconv"
	"workshop/internal/todo"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) Publish(ctx context.Context, event todo.Event) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(event.Todo.Id, 10)),
		Value: eventBytes,
	})
}

```

# Kafka реализация Consumer

Установим зависимости

```
go get github.com/sirupsen/logrus
```

Создадим файл `internal/todo/kafka/consumer.go` :

```go
package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	reader *kafka.Reader
	logger logrus.FieldLogger
}

func NewConsumer(reader *kafka.Reader, logger logrus.FieldLogger) *Consumer {
	return &Consumer{reader: reader, logger: logger}
}

func (c Consumer) Consume() error {
	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		c.logger.WithFields(map[string]interface{}{
			"topic": msg.Topic,
			"key": string(msg.Key),
			"value": string(msg.Value),
			"partition": msg.Partition,
			"time": msg.Time.String(),
		}).Info("Message received")
	}
}
```

# Пишем бизнес-логику Todo

Определим интерфейс `TodoService`  в `internal/todo/http/handler.go`:

```go
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
```

Напишем реализацию под этот интерфейс. Внутри файла `internal/todo/service.go`:

```go
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
```

# Напишем хендлеры для HTTP запросов

Добавим зависимость

```
go get github.com/gin-gonic/gin
```

Создадим врапперы для ответа ошибки внутри файла `internal/todo/http/response.go`:

```go
package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"workshop/internal/todo"
)

var httpStatusCodeMap = map[int]int{
	todo.ErrNotFound:     http.StatusNotFound,
	todo.ErrValidation:   http.StatusBadRequest,
	todo.ErrEventPublish: http.StatusInternalServerError,
	todo.ErrInternal:     http.StatusInternalServerError,
}

type Error struct {
	Msg string `json:"msg"`
}

func NewError(ctx *gin.Context, err error) {
	var todoError *todo.Error
	if errors.As(err, &todoError) {
		status, ok := httpStatusCodeMap[todoError.Code]
		if !ok {
			status = http.StatusInternalServerError
		}

		ctx.JSON(status, Error{Msg: todoError.Msg})
		return
	}

	ctx.JSON(http.StatusInternalServerError, Error{Msg: todoError.Msg})
}

func BadRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, Error{Msg: err.Error()})
}
```

Создадим структуры реквестов для create + update внутри `internal/todo/http/request.go` файла:

```go
package http

type CreateTodo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
```

Создадим структуру хендлера в файле `internal/todo/http/handler.go`

```go
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
```

Добавим роуты в файл `internal/todo/http/routes.go`:

```go
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"workshop/internal/todo"
	todokafka "workshop/internal/todo/kafka"
	"workshop/internal/todo/mysql"
)

type RouteProvider struct {
	db       *sqlx.DB
	producer *kafka.Writer
	router   *gin.Engine
}

func NewRouteProvider(db *sqlx.DB, producer *kafka.Writer, router *gin.Engine) *RouteProvider {
	return &RouteProvider{db: db, producer: producer, router: router}
}

func (routeProvider *RouteProvider) Routes() {
	publisher := todokafka.NewPublisher(routeProvider.producer)
	repository := mysql.NewRepository(routeProvider.db)
	todoService := todo.NewService(repository, publisher)
	handler := NewTodoHandler(todoService)

	routeProvider.router.GET("/todos", handler.GetAll)
	routeProvider.router.POST("/todos", handler.Create)
	routeProvider.router.GET("/todos/:id", handler.GetById)
	routeProvider.router.PUT("/todos/:id", handler.Update)
	routeProvider.router.DELETE("/todos/:id", handler.Delete)
}
```

# Server

Внутри `cmd/server` создадим файл `main.go`. Внутри него напишем код для запуска приложения в виде http сервера.
В `main.go` мы будем:

- Загружать переменные окружения из `.env`
- Инициализировать логгер
- Подключаться к БД
- Запускать миграции
- Создавать подключение к Kafka
- Добавлять роуты для todo модуля.
- Слушать http подключения

```
go get github.com/joho/godotenv
go get github.com/golang-migrate/migrate
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

Файл  `main.go`:

```go
package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"os"
	"strings"
	"time"
	_ "workshop/docs"
	todo "workshop/internal/todo/http"
)

func main() {
	//init logger
	logger := logrus.New()

	//load env vars from .env
	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	//connect to db
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME")))
	if err != nil {
		logger.Fatal(err)
	}

	//init migrations driver
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{
		MigrationsTable: "migrations",
		DatabaseName:    os.Getenv("DB_NAME"),
	})
	if err != nil {
		logger.Fatal(err)
	}

	//init migrations instance
	m, err := migrate.NewWithDatabaseInstance("file://migrations", os.Getenv("DB_NAME"), driver)
	if err != nil {
		logger.Fatal(err)
	}

	//run migrations
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal(err)
	}

	//init kafka writer
	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(os.Getenv("KAFKA_BROKERS"), ",")...),
		Topic:        os.Getenv("TODO_TOPIC"),
		Balancer:     &kafka.Hash{},
		BatchTimeout: time.Millisecond,
		BatchSize:    1,
	}
	defer w.Close()

	//init router
	router := gin.Default()

	// use ginSwagger middleware to serve the API docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//add todo module routes
	todo.NewRouteProvider(db, w, router).Routes()

	//listen and serve http requests
	if err := router.Run(":" + os.Getenv("HTTP_PORT")); err != nil {
		logrus.Error(err)
	}
}

```

# Добавим swagger описание для todo routes

Чтобы сгенерировать документацию:

```bash
swag init --dir=cmd/server --parseInternal --pd
```

генерация может занять некоторое время.

Сейчас документация пустая, чтобы там появились наши ендпоинты, мы должно добавить комментарии к нашим хендлерам.

Внутри файла `internal/todo/http/handler.go` добавим комментарии под каждым методом, который отвечает обработку http запросов (`Create`, `GetAll`, `GetById`, `Update`, `Delete`).

Над методом `Create` добавим:

```go
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
```

Над методом `GetAll` добавим:

```go
// GetAllTodos godoc
// @Summary      Get all todos
// @Description  Get all todos
// @Tags         todo
// @Accept       json
// @Produce      json
// @Success      200  {array}  todo.Todo
// @Failure      500  {object}  Error
// @Router       /todos [get]
```

Над методом `GetById` добавим:

```go
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
```

Над методом `Update` добавим:

```go
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
```

Над методом `Delete` добавим:

```go
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
```

Теперь снова генерируем сваггер доку командой выше, запускаем сервер

```go
go run cmd/server/main.go
```

и в браузере открываем [http://localhost:7777/swagger/index.html](http://localhost:7777/swagger/index.html)

# Consumer

Создадим файл `cmd/consumer/main.go`:

```go
package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	todo "workshop/internal/todo/kafka"
)

func main() {
	//init logger
	logger := logrus.New()

	//load env vars from .env
	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	//init reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		Topic:   os.Getenv("TODO_TOPIC"),
		GroupID: "todo-consumer",
	})

	//create todo consumer
	todoConsumer := todo.NewConsumer(reader, logger)
	logger.Info("start consuming messages from kafka")

	//start consuming
	err := todoConsumer.Consume()
	if err != nil {
		logger.Error("error during reading msges from kafka:", err)
	}
}
```

Запустим командой:

```bash
go run cmd/consumer/main.go
```