package routing

import (
	"net/http"

	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
)

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
