package filesystem

import (
	"embly/pkg/tester"
	"testing"
)

func TestZipSources(te *testing.T) {
	t := tester.New(te)

	projectRoot, err := tester.CreateTmpTestProject()
	t.PanicOnErr(err)

	_ = projectRoot
}
