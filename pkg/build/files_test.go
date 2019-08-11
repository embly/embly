package build

import (
	"bytes"
	"testing"

	rc "embly/pkg/rustcompile/proto"
	"embly/pkg/tester"
)

var plainMain = &rc.File{
	Path: "./main.rs",
	Body: []byte(`fn main() {
		println!("hello world")
	}`),
}

func minimalOneFile() (pf projectFiles) {
	return newProjectFiles([]*rc.File{plainMain})
}

func withMisplacedToml() (pf projectFiles) {
	return newProjectFiles([]*rc.File{
		plainMain,
		&rc.File{
			Path: "./foo/Cargo.toml",
			Body: []byte(`fn main() {
			println!("hello world")}`),
		},
	})
}

// func

func TestBasicValidateAndClean(te *testing.T) {
	t := tester.New(te)

	pf := minimalOneFile()
	t.AssertNoError(pf.validateAndClean())

	if !bytes.Equal(pf.files["Cargo.toml"].Body, defaultCargoToml) {
		t.Error("cargo.toml isn't correct")
	}

	pf = withMisplacedToml()
	t.AssertNoError(pf.validateAndClean())
	if !bytes.Equal(pf.files["Cargo.toml"].Body, defaultCargoToml) {
		t.Error("cargo.toml isn't correct")
	}

}
