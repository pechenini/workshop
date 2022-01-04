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

func newError(ctx *gin.Context, err error) {
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

func badRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, Error{Msg: err.Error()})
}