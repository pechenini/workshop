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
