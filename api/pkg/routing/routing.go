package routing

import (
	"net/http"

	"context"
	"database/sql"

	rc "embly/api/pkg/rustcompile/proto"
	"embly/api/pkg/cache"

	"github.com/gin-gonic/gin"
)

// Context is the context passed to a request
type Context struct {
	DB       *sql.DB
	RCClient rc.RustCompileClient
	RedisClient cache.Client
}

// ErrorWrapHandler ...
func ErrorWrapHandler(rc *Context,
	handler func(context.Context, *Context,
		*gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := context.Background()
		err := handler(ctx, rc, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		}
	}
}
