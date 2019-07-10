package build

import (
	"context"
	"io/ioutil"
	"mime/multipart"
	"strings"

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
	if err = c.Request.ParseMultipartForm(4 * 1000 * 1000); err != nil {
		err = errors.Wrap(err, "error parsing multipart form")
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
				err = errors.Wrap(err, "error opening mpf file "+mpf.Filename)
				return
			}
			var b []byte
			if b, err = ioutil.ReadAll(f); err != nil {
				err = errors.Wrap(err, "error reading all of file "+mpf.Filename)
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

func buildHandler(ctx context.Context, roc *routing.Context, c *gin.Context) (err error) {
	name, files, err := parseFormAndGenFiles(c)
	if err != nil {
		// TODO: 429 err
		return err
	}
	pf := newProjectFiles(files)
	if err = pf.validateAndClean(); err != nil {
		err = errors.Wrap(err, "error validating and cleaning files")
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		err = errors.Wrap(err, "error generating uuid")
		return
	}
	fun := models.Function{
		ID:   id.String(),
		Name: name,
	}
	if err = fun.Insert(ctx, roc.DB, boil.Infer()); err != nil {
		return err
	}

	sbClient, err := roc.RCClient.StartBuild(ctx, pf.toCode())
	if err != nil {
		return err
	}
	var logOutput strings.Builder
	var result *rc.Result
	var resultingBinary []byte
	for {
		if result, err = sbClient.Recv(); err != nil {
			return err
		}
		if result.Stdout != nil {
			logOutput.Write(result.Stdout)
		}
		if result.Stderr != nil {
			logOutput.Write(result.Stderr)
		}
		if len(result.Binary) != 0 {
			resultingBinary = result.Binary
			break
		}
	}

	if err = roc.RedisClient.SetB(fun.ID, resultingBinary); err != nil {
		return
	}

	c.JSON(200, gin.H{"function": gin.H{"id": fun.ID, "name": fun.Name}})

	return nil
}
