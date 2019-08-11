package build

import (
	"bytes"
	"database/sql/driver"
	"embly/pkg/routing"
	"embly/pkg/tester"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	rc "embly/pkg/rustcompile/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func handler(rc *routing.Context) http.Handler {
	r := gin.Default()
	ApplyRoutes(rc, r.Group("/"))
	return r
}

type AnyValue struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyValue) Match(v driver.Value) bool {
	return true
}

func TestBuildWithFiles(te *testing.T) {
	t := tester.New(te)
	roc, mock, rctc := t.NewRoutingContext()
	s := httptest.NewServer(handler(roc))
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
		AnyValue{},
	).WillReturnResult(sqlmock.NewResult(1, 1))

	rctc.ResultChan <- &rc.Result{
		Stdout: []byte("hi"),
	}
	rctc.ResultChan <- &rc.Result{
		Binary: []byte("hi"),
	}

	resp, _ := t.AssertCode(http.StatusOK)(client.Do(req))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
	var br map[string]buildResp
	json.Unmarshal(b, &br)
	bb, _ := roc.RedisClient.GetB(br["function"].ID)
	t.Assert().Equal(bb, []byte("hi"))
}

type buildResp struct {
	ID   string
	Name string
}
