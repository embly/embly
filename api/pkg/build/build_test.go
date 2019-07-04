package build

import (
	"bytes"
	"context"
	"database/sql"
	"embly/api/pkg/config"
	"embly/api/pkg/dbutil"
	"embly/api/pkg/routing"
	"embly/api/pkg/tester"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

func init() {
	config.Register(
		"DB_HOST",
		"DB_PORT",
		"DB_USERNAME",
		"DB_DATABASE",
		"DB_PASSWORD",
		"REDIS_HOST",
	)
	var err error
	db, err = dbutil.Connect()
	if err != nil {
		log.Fatal(err)
	}
}

func handler() http.Handler {
	r := gin.Default()
	r.POST("/", routing.ErrorWrapHandler(db, parseFormAndGenFilesTestHandler))
	return r
}

func parseFormAndGenFilesTestHandler(ctx context.Context, db *sql.DB, c *gin.Context) (err error) {
	_, err = parseFormAndGenFiles(c)
	if err != nil {
		return err
	}
	c.JSON(200, gin.H{"msg": "ok"})
	return nil
}

func TestBuildWithFiles(te *testing.T) {
	t := tester.New(te)
	s := httptest.NewServer(handler())

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("whatever", "./src/main.rs")
	t.AssertNoError(err)
	part.Write([]byte(`fn main(){ println!("hi") }`))

	part, err = writer.CreateFormFile("foo", "./Cargo.toml")
	t.AssertNoError(err)
	part.Write([]byte(`[dependencies]
embly="*"
	`))

	t.AssertNoError(writer.Close())
	req, err := http.NewRequest("POST", s.URL, body)
	t.AssertNoError(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	t.AssertCode(http.StatusOK)(client.Do(req))

}
