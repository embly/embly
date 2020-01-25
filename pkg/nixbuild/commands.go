package nixbuild

import (
	"embly/pkg/config"
	"embly/pkg/embly"
	"embly/pkg/filesystem"
	_ "embly/pkg/nixbuild/statik"
	"fmt"

	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
)

func CleanAllDependencies() (err error) {
	emblyDir, err := embly.EmblyDir()
	if err != nil {
		return
	}
	dirs, err := filepath.Glob(filepath.Join(emblyDir, "./nix/") + "*")
	if err != nil {
		return
	}
	for _, file := range dirs {
		if err = os.RemoveAll(file); err != nil {
			return
		}
	}
	return
}

func BuildRust() (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	emblyDir, err := embly.EmblyDir()
	if err != nil {
		return
	}

	cfg, err := config.New("../../examples/hello-world")
	if err != nil {
		return
	}

	if err = DownloadDependencies([]string{"rust", "lucetc"}); err != nil {
		return
	}

	project := filesystem.NewProject(cfg)
	dir, err := project.CopyFunctionSourcesToTmp("hello")
	if err != nil {
		return
	}
	fmt.Println(dir)
	cmd := exec.Command(
		filepath.Join(emblyDir, "./nix/rust/bin/cargo"),
		"build",
		"--target=wasm32-wasi",
		fmt.Sprintf("--manifest-path=%s/Cargo.toml", dir), // TODO: support function location within subdir
		fmt.Sprintf("--target-dir=%s", filepath.Join(emblyDir, "./rust-target")),
	)
	cmd.Env = []string{
		fmt.Sprintf("RUSTC=%s", filepath.Join(emblyDir, "./nix/rust/bin/rustc")),
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func DownloadDependencies(items []string) (err error) {
	if err = embly.CreateHomeDir(); err != nil {
		return err
	}
	emblyDir, err := embly.EmblyDir()
	if err != nil {
		return
	}
	for _, item := range items {
		if _, err = os.Stat(filepath.Join(emblyDir, "./nix/", item)); err == nil {
			// directory exists, continue to other dependencies
			continue
		}
		if err = BuildFile(item); err != nil {
			return
		}
	}
	return
}

func BuildFile(name string) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()

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
