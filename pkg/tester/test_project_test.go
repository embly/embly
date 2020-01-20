package tester

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultProjectCreation(te *testing.T) {
	t := New(te)
	loc, err := CreateTmpTestProject()
	t.PanicOnErr(err)
	for _, fileMap := range []map[string]string{
		ExternalDependency,
		ExternalDependency2,
		EmblyFile,
		BasicRustProject,
		FooRustProject,
	} {
		for name := range fileMap {
			_, err := os.Stat(filepath.Join(loc, name))
			t.PanicOnErr(err, name)
		}
	}
}
