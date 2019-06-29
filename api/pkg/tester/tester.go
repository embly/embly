package tester

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Tester struct {
	*testing.T
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
