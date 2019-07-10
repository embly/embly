package tester

import (
	"context"
	"database/sql"
	"embly/api/pkg/routing"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	rc "embly/api/pkg/rustcompile/proto"
	"embly/api/pkg/cache"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// Tester is a wrapper around testing.T
type Tester struct {
	*testing.T
}

// RustCompileTestClient is a test client for RustCompileClient, it provoides a channel to assert on re.Code values
type RustCompileTestClient struct {
	CodeChan   chan (*rc.Code)
	ResultChan chan (*rc.Result)
	ErrChan    chan (error)
}

// StartBuild just takes the code value and returns it on a channel
func (c *RustCompileTestClient) StartBuild(ctx context.Context, in *rc.Code, opts ...grpc.CallOption) (sbc rc.RustCompile_StartBuildClient, err error) {
	c.CodeChan <- in
	rctsbc := rustCompileTestStartBuildClient{
		ResultChan: c.ResultChan,
		ErrChan:    c.ErrChan,
	}
	return &rctsbc, err
}

type rustCompileTestStartBuildClient struct {
	grpc.ClientStream
	ResultChan chan (*rc.Result)
	ErrChan    chan (error)
}

func (x *rustCompileTestStartBuildClient) Recv() (r *rc.Result, err error) {
	select {
	case r = <-x.ResultChan:
	case err = <-x.ErrChan:
	default:
	}
	return
}

// newRustCompileTestClient creates a new RustCompileTestClient as an rc.RustCompileClient interface
func newRustCompileTestClient() RustCompileTestClient {
	return RustCompileTestClient{
		CodeChan:   make(chan *rc.Code, 100),
		ResultChan: make(chan *rc.Result, 100),
		ErrChan:    make(chan error, 100),
	}
}

// NewRoutingContext creates a routing context for tests. it has a mocked db
func (t *Tester) NewRoutingContext() (rc *routing.Context, mock sqlmock.Sqlmock, rctc RustCompileTestClient) {
	var db *sql.DB
	db, mock = t.NewMockDB()
	rctc = newRustCompileTestClient()
	rc = &routing.Context{
		DB:          db,
		RCClient:    &rctc,
		RedisClient: cache.NewTestClient(),
	}
	return
}

// NewMockDB creates a new sqlmock db and mock object
func (t *Tester) NewMockDB() (db *sql.DB, mock sqlmock.Sqlmock) {
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return
}

// New wraps a *testing.T
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
