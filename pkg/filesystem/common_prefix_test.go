package filesystem

import (
	"embly/pkg/tester"
	"testing"
)

func TestCommonBasic(te *testing.T) {
	t := tester.New(te)
	t.Assert().Equal(CommonPrefix([]string{
		"/foo/bar/baz/",
		"/foo/bar/bar/",
	}), "/foo/bar/")

	t.Assert().Equal(CommonPrefix([]string{
		"/foo/bar/baz/",
		"/foo/bar/bazolemule/",
	}), "/foo/bar/")

	t.Assert().Equal(CommonPrefix([]string{
		"/foo/bar/baz/",
	}), "/foo/bar/baz/")

	t.Assert().Equal(CommonPrefix([]string{}), "/")

	t.Assert().Equal(CommonPrefix([]string{"/no", "/yes"}), "/")

}
