package embly

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// EmblyDir gets the location of the embly directory
func EmblyDir() (dir string, err error) {
	home, err := homedir.Dir()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	dir = filepath.Join(home, "./.embly/")
	return
}

// CreateHomeDir creates the embly directory in the users home directory
func CreateHomeDir() (err error) {
	dir, err := EmblyDir()
	if err != nil {
		return
	}
	for _, folder := range []string{
		"./", "./nix",
		"./blob_cache",
		"./build_context",
		"./build_context/rust_target",
		"./build_context/cargo_home",
		"./lucet_cache", "./build",
		"./result",
	} {
		loc := filepath.Join(dir, folder)
		_, err = os.Stat(loc)
		if err != nil {
			err = os.MkdirAll(loc, os.ModePerm)
			if err != nil {
				err = errors.WithStack(err)
				return err
			}
		}
	}
	return nil
}
