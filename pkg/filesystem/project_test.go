package filesystem

import (
	"embly/pkg/config"
	"embly/pkg/tester"
	"fmt"
	"os"
	"path/filepath"
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

	project, err := NewProject(cfg)
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
