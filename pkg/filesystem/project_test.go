package filesystem

import (
	"embly/pkg/config"
	"embly/pkg/tester"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProject(te *testing.T) {
	// TODO: inject debounceTime and tick time override to make this test
	// run faster

	t := tester.New(te)

	loc, err := tester.CreateTmpTestProject()
	t.PanicOnErr(err)

	cfg, err := config.New(loc)
	t.PanicOnErr(err)

	project := NewProject(cfg)
	t.PanicOnErr(err)
	t.PanicOnErr(project.Start())

	start := time.Now()
	var eventTime time.Time
	var lock sync.Mutex
	var gotFunctions []config.Function
	go project.NextEvent(func(fn config.Function) {
		fmt.Println(fn)
		eventTime = time.Now()
		lock.Lock()
		gotFunctions = append(gotFunctions, fn)
		lock.Unlock()
	})
	t.PanicOnErr(os.Chtimes(filepath.Join(loc, "../lib/Cargo.toml"), time.Now(), time.Now()))
	time.Sleep(debounceTime - time.Second)
	t.PanicOnErr(os.Chtimes(filepath.Join(loc, "../lib/Cargo.toml"), time.Now(), time.Now()))
	time.Sleep(debounceTime + time.Second)

	lock.Lock()
	t.Assert().ElementsMatch(gotFunctions, cfg.Functions)
	lock.Unlock()
	t.Assert().True(eventTime.Sub(start) > debounceTime+(debounceTime/2))
}

func TestProjectCopy(te *testing.T) {
	t := tester.New(te)
	loc, err := tester.CreateTmpTestProject()
	t.PanicOnErr(err)
	cfg, err := config.New(loc)
	t.PanicOnErr(err)
	project := NewProject(cfg)
	dir, err := project.CopyFunctionSourcesToBuild("", "foo")
	// we need the path of the root, not the function path
	dir = filepath.Join(dir, "../../")

	t.PanicOnErr(err)
	files := []string{}
	t.PanicOnErr(filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		files = append(files, path)
		return err
	}))
	prefix := CommonPrefix(files)
	trimmed := []string{}
	//                        ignore the parent dir
	for _, file := range files[1:] {
		trimmed = append(trimmed, strings.TrimPrefix(file, prefix))
	}
	t.Assert().Equal(trimmed, []string{
		"lib",
		"lib/Cargo.toml",
		"lib/src",
		"lib/src/lib.rs",
		"lib2",
		"lib2/Cargo.toml",
		"lib2/src",
		"lib2/src/lib.rs",
		"project",
		"project/foo",
		"project/foo/Cargo.toml",
		"project/foo/src",
		"project/foo/src/main.rs"})
}
