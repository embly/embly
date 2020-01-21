package nixbuild

import (
	"embly/pkg/tester"
	"testing"
)

func TestNixBuild(te *testing.T) {
	t := tester.New(te)
	t.PanicOnErr(BuildFile("rust"))
}
