package nixbuild

import (
	"embly/pkg/embly"
	_ "embly/pkg/nixbuild/statik"
	"fmt"

	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
)

func BuildFile(name string) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()

	if err = embly.CreateHomeDir(); err != nil {
		return err
	}

	staticFiles, err := fs.New()
	if err != nil {
		return
	}
	f, err := staticFiles.Open(fmt.Sprintf("/%s.nix", name))
	if err != nil {
		return
	}

	cmd := exec.Command("nix-build", "-", "-o", name)
	emblyDir, err := embly.EmblyDir()
	if err != nil {
		return
	}
	cmd.Dir = filepath.Join(emblyDir, "./nix")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	if _, err = io.Copy(stdin, f); err != nil {
		return
	}
	stdin.Close()

	if err = cmd.Start(); err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}

	return nil

}
