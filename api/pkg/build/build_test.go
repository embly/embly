package build

import (
	"bytes"
	"database/sql"
	"embly/api/pkg/config"
	"embly/api/pkg/dbutil"
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
	g := r.Group("/")
	ApplyRoutes(db, g)
	return r
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
