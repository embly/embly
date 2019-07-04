package routing

import (
	"net/http"

	"context"
	"database/sql"

	rc "embly/api/pkg/rustcompile/proto"

	"github.com/gin-gonic/gin"
)

// Context is the context passed to a request
type Context struct {
	DB       *sql.DB
	RCClient *rc.RustCompileClient
}

// ErrorWrapHandler ...
func ErrorWrapHandler(db *sql.DB,
	handler func(context.Context, *sql.DB,
		*gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := context.Background()
		err := handler(ctx, db, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		}
	}
}
