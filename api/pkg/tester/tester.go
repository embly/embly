package tester

import (
	"database/sql"
	"embly/api/pkg/routing"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Tester is a wrapper around testing.T
type Tester struct {
	*testing.T
}

func (t *Tester) NewRoutingContext() (rc *routing.Context, mock sqlmock.Sqlmock) {
	var db *sql.DB
	db, mock = t.NewMockDB()
	rc = &routing.Context{
		DB:       db,
		RCClient: nil, //TODO
	}
	return
}

func (t *Tester) NewMockDB() (db *sql.DB, mock sqlmock.Sqlmock) {
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return
}

func New(t *testing.T) Tester {
	return Tester{t}
}

func (t *Tester) AssertNoError(errs ...interface{}) {
	if len(errs) > 0 && errs[0] != nil {
		t.Error(errs...)
	}
}

// Assert returns a testify/assert instance
func (t *Tester) Assert() *assert.Assertions {
	return assert.New(t)
}

func (t *Tester) AssertCode(code int) func(*http.Response, error) (*http.Response, error) {
	return func(resp *http.Response, err error) (*http.Response, error) {
		t.Assert().NoError(err)
		if resp.StatusCode != code {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Assert().Equal(resp.StatusCode, code,
				fmt.Sprintf("Expected status code %d got status code %d: %s", code, resp.StatusCode, string(b)))
		}
		return resp, err
	}
}

func (t *Tester) AssertContains(substring string) func(*http.Response, error) (*http.Response, error) {
	return func(resp *http.Response, err error) (*http.Response, error) {
		b, err := ioutil.ReadAll(resp.Body)
		t.Assert().NoError(err)
		t.Assert().Contains(string(b), substring, fmt.Sprintf("Body %#v doesn't contain substring %#v", string(b), substring))
		return resp, err
	}
}
