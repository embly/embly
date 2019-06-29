package build

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"

	"embly/api/pkg/routing"

	rc "embly/api/pkg/rustcompile/proto"

	"github.com/gin-gonic/gin"
)

// ApplyRoutes will apply calendar routes
func ApplyRoutes(db *sql.DB, r *gin.RouterGroup) {
	// boil.DebugMode = true
	r.GET("/", routing.ErrorWrapHandler(db, indexHandler))
	r.POST("/", routing.ErrorWrapHandler(db, buildHandler))
}

func indexHandler(ctx context.Context, db *sql.DB, c *gin.Context) error {
	c.JSON(200, gin.H{"msg": "Hello"})
	return nil
}

func buildHandler(ctx context.Context, db *sql.DB, c *gin.Context) error {
	err := c.Request.ParseMultipartForm(10)
	if err != nil {
		return err
	}
	files := []*rc.File{}
	for _, mpfs := range c.Request.MultipartForm.File {
		for _, mpf := range mpfs {
			f, err := mpf.Open()
			if err != nil {
				return err
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			files = append(files, &rc.File{
				Path: mpf.Filename,
				Body: b,
			})
		}
	}
	pf := newProjectFiles(files)
	if err := pf.validateAndClean(); err != nil {
		return err
	}

	fmt.Println(pf.toCode())
	c.JSON(200, gin.H{"msg": "Hello"})

	return nil
}
