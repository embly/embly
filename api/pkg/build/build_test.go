package build

import (
	"bytes"
	"database/sql/driver"
	"embly/api/pkg/routing"
	"embly/api/pkg/tester"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func handler(rc *routing.Context) http.Handler {
	r := gin.Default()
	r.POST("/", routing.ErrorWrapHandler(rc, buildHandler))
	return r
}

type AnyValue struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyValue) Match(v driver.Value) bool {
	return true
}

func TestBuildWithFiles(te *testing.T) {
	t := tester.New(te)
	rc, mock := t.NewRoutingContext()
	s := httptest.NewServer(handler(rc))
	_ = mock
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "foo")
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

	mock.ExpectExec(`INSERT INTO "functions"`).WithArgs(
		AnyValue{}, "foo",
		AnyValue{}, AnyValue{},
		AnyValue{}, AnyValue{},
	).WillReturnResult(sqlmock.NewResult(1, 1))

	t.AssertCode(http.StatusOK)(client.Do(req))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
