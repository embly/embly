package dock

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

func newClient(t *testing.T) *Client {
	c, err := NewClient()
	if err != nil {
		t.Fatal(errors.Wrap(err, "error creating client"))
	}
	return c

}

func TestVolumeExists(t *testing.T) {
	c := newClient(t)
	fmt.Println(c.VolumeExists("vinyl_ivycache"))
	fmt.Println(c.VolumeExists("vinyl_ivyce"))
	fmt.Println(c.CreateVolume("foo"))

	fmt.Println(StartVinyl("hi"))
}

func TestCreate(t *testing.T) {
	t.Skip()

	wd, _ := os.Getwd()

	buildLocation := filepath.Join(wd, "../../examples/project/hello")
	destination := filepath.Join(wd, "out.out")

	if err := CompileRust(CompileRustSettings{
		FunctionName:   "hello",
		BuildLocation:  buildLocation,
		DestinationDir: destination,
	}); err != nil {
		fmt.Printf("%+v", err)
		t.Error(err)
	}

	// defer os.Remove(destination)
	info, err := os.Stat(destination)
	if err != nil {
		t.Error(err)
	}
	_ = info

}

func TestImageExists(t *testing.T) {
	t.Skip()
	c := newClient(t)

	if exists, _ := c.ImageExists("foo"); exists == true {
		t.Error("shouldn't exist")
	}
}

func TestImageAge(t *testing.T) {
	t.Skip()
	c := newClient(t)
	_, _ = c.ImageCreated("embly/compile-rust-wasm")
	_, _ = c.ImageCreated("qoomon/docker-host")

}

func TestPullImage(t *testing.T) {
	t.Skip()
	c := newClient(t)
	_ = c.PullImage("embly/vinyl")
	_ = c.PullImage("python:3-slim")
}
