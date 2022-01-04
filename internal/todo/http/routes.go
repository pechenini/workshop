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