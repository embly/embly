package build

import (
	"context"
	"database/sql"
	"io/ioutil"

	"embly/api/pkg/models"
	"embly/api/pkg/routing"

	rc "embly/api/pkg/rustcompile/proto"

	"github.com/getlantern/uuid"
	"github.com/gin-gonic/gin"
	"github.com/volatiletech/sqlboiler/boil"
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

func parseFormAndGenFiles(c *gin.Context) (files []*rc.File, err error) {
	if err = c.Request.ParseMultipartForm(10); err != nil {
		return
	}
	for _, mpfs := range c.Request.MultipartForm.File {
		for _, mpf := range mpfs {
			f, err := mpf.Open()
			if err != nil {
				return files, err
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				return files, err
			}
			files = append(files, &rc.File{
				Path: mpf.Filename,
				Body: b,
			})
		}
	}

	return
}

func buildHandler(ctx context.Context, db *sql.DB, c *gin.Context) error {
	files, err := parseFormAndGenFiles(c)
	if err != nil {
		return err
	}
	pf := newProjectFiles(files)
	if err := pf.validateAndClean(); err != nil {
		return err
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	fun := models.Function{
		ID: id.String(),
	}
	fun.Insert(ctx, db, boil.Infer())
	// pf.toCode()

	c.JSON(200, gin.H{"function": gin.H{"id": fun.ID}})

	return nil
}
