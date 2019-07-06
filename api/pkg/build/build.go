package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"

	"embly/api/pkg/models"
	"embly/api/pkg/routing"

	rc "embly/api/pkg/rustcompile/proto"

	"github.com/getlantern/uuid"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

// ApplyRoutes will apply calendar routes
func ApplyRoutes(rc *routing.Context, r *gin.RouterGroup) {
	boil.DebugMode = true
	r.GET("/", routing.ErrorWrapHandler(rc, indexHandler))
	r.POST("/", routing.ErrorWrapHandler(rc, buildHandler))
}

func indexHandler(ctx context.Context, rc *routing.Context, c *gin.Context) error {
	c.JSON(200, gin.H{"msg": "Hello"})
	return nil
}

func parseFormAndGenFiles(c *gin.Context) (name string, files []*rc.File, err error) {
	if err = c.Request.ParseMultipartForm(10); err != nil {
		return
	}
	for key, values := range c.Request.MultipartForm.Value {
		if key == "name" {
			name = values[len(values)-1]
		}
	}
	if name == "" {
		err = errors.New("Build upload must have a name")
		return
	}
	for _, mpfs := range c.Request.MultipartForm.File {
		for _, mpf := range mpfs {
			var f multipart.File
			if f, err = mpf.Open(); err != nil {
				return
			}
			var b []byte
			if b, err = ioutil.ReadAll(f); err != nil {
				return
			}
			files = append(files, &rc.File{
				Path: mpf.Filename,
				Body: b,
			})
		}
	}

	return
}

func buildHandler(ctx context.Context, rc *routing.Context, c *gin.Context) error {
	name, files, err := parseFormAndGenFiles(c)
	if err != nil {
		// TODO: 429 err
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
		ID:   id.String(),
		Name: name,
	}
	if err = fun.Insert(ctx, rc.DB, boil.Infer()); err != nil {
		return err
	}
	fmt.Println(pf.toCode())

	c.JSON(200, gin.H{"function": gin.H{"id": fun.ID}})

	return nil
}
