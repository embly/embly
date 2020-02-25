package nixbuild

import (
	"embly/pkg/config"
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

	cfg, err := config.New("../../examples/kv")
	if err != nil {
		return
	}

	builder, err := NewBuilder()
	t.PanicOnErr(err)

	t.PanicOnErr(builder.writeNixFiles())

	builder.SetProject(cfg)
	t.PanicOnErr(builder.CleanAllDependencies())

	result, err := builder.BuildFunction("main")
	t.PanicOnErr(err)
	t.PanicOnErr(os.RemoveAll(result))

}
