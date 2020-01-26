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
	builder, err := NewBuilder()
	t.PanicOnErr(err)
	t.PanicOnErr(builder.CleanAllDependencies())
	t.PanicOnErr(builder.BuildRust())
}
