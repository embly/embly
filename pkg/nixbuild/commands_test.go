package nixbuild

import (
	"embly/pkg/tester"
	"os"
	"testing"
)

func TestNixBuild(te *testing.T) {
	t := tester.New(te)
	if os.Getenv("NIXBUILD_INTEGRATION_TEST") == "" {
		t.Skip()
		return
	}
	t.PanicOnErr(CleanAllDependencies())
	t.PanicOnErr(BuildRust())
}
