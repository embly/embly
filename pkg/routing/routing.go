package routing

import (
	"fmt"
	"net/http"

	"context"
	"database/sql"
	"embly/api/pkg/cache"
	"embly/api/pkg/config"
	"embly/api/pkg/dbutil"
	"embly/api/pkg/rustcompile"
	rc "embly/api/pkg/rustcompile/proto"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

// Context is the context passed to a request
type Context struct {
	DB          *sql.DB
	RCClient    rc.RustCompileClient
	RedisClient cache.Client
}

// NewContext ...
func NewContext() (c *Context, err error) {
	config.Register(
		"DB_HOST",
		"DB_DATABASE",
		"DB_PASSWORD",
		"DB_PORT",
		"DB_USERNAME",
		"RC_HOST",
		"REDIS_HOST",
	)
	c = &Context{}
	if c.DB, err = dbutil.Connect(); err != nil {
		return
	}
	if c.RCClient, err = rustcompile.NewRustCompileClient(config.Get("RC_HOST")); err != nil {
		return
	}
	var rconn redis.Conn
	if rconn, err = redis.Dial("tcp", config.Get("REDIS_HOST")); err != nil {
		return
	}
	c.RedisClient = cache.NewClient(rconn)

	return
}

// ErrorWrapHandler ...
func ErrorWrapHandler(rc *Context,
	handler func(context.Context, *Context,
		*gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := context.Background()
		err := handler(ctx, rc, c)
		if err != nil {
			stack := fmt.Sprintf("%+v", err)
			fmt.Println(stack)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": stack})
		}
	}
}
