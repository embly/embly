package localbuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	t.Skip()

	wd, _ := os.Getwd()

	buildLocation := filepath.Join(wd, "../../examples/project/hello")
	buildContext := filepath.Join(wd, "../../")
	destination := filepath.Join(wd, "out.out")

	if err := Create("hello", buildLocation, buildContext, destination); err != nil {
		t.Error(err)
	}

	// defer os.Remove(destination)
	info, err := os.Stat(destination)
	if err != nil {
		t.Error(err)
	}
	_ = info

}
