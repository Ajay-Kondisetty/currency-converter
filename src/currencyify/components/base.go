package components

import (
	"context"

	"currencyify/utils"

	"github.com/gomodule/redigo/redis"
)

type BaseComponent struct {
	ReqCtx    context.Context
	AppError  *utils.AppError
	RedisConn redis.Conn
}

var ComponentMap = make(map[string]func(*BaseComponent) interface{})
