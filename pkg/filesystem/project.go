package filesystem

import (
	"crypto/sha256"
	"embly/pkg/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

type trackedFile struct {
	fi   os.FileInfo
	hash []byte
}

type Project struct {
	watcher           *watcher.Watcher
	cfg               *config.Config
	functionLocations map[string][]config.Function

	notify chan config.Function

	fnMap      map[string]config.Function
	fnTimerMap map[string]*time.Timer

	files map[string]*trackedFile
}

// NewProject should create a new project
func NewProject(cfg *config.Config) (p *Project, err error) {
	p = &Project{
		watcher:           watcher.New(),
		cfg:               cfg,
		notify:            make(chan config.Function, 100),
		fnMap:             map[string]config.Function{},
		fnTimerMap:        map[string]*time.Timer{},
		functionLocations: map[string][]config.Function{},
		files:             map[string]*trackedFile{},
	}

	// if functions depend on shared sources watcher.Watcher will de-dupe them
	for _, fn := range cfg.Functions {
		if err = p.AddRecursive(fn.Path, fn); err != nil {
			return
		}
		for _, source := range fn.Sources {
			if err = p.AddRecursive(source, fn); err != nil {
				return
			}
		}
		p.fnMap[fn.Name] = fn
	}

	for name, fi := range p.watcher.WatchedFiles() {
		tracked := &trackedFile{
			fi: fi,
		}
		if !fi.IsDir() {
			var f *os.File
			f, err = os.Open(name)
			if err != nil {
				return
			}
			h := sha256.New()
			if _, err = io.Copy(h, f); err != nil {
				return
			}
			tracked.hash = h.Sum(nil)
		}
		p.files[name] = tracked
	}

	return
}

func (p *Project) AddRecursive(path string, function config.Function) (err error) {
	path = p.cfg.AbsolutePath(path)
	if err = p.watcher.Add(path); err != nil {
		return err
	}
	if err = filepath.Walk(path, func(thisPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Name() == "target" && fi.IsDir() {
			return filepath.SkipDir
		}
		if fi.IsDir() {
			// the lib handles adding children
			if err := p.watcher.Add(thisPath); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return
	}
	p.functionLocations[path] = append(p.functionLocations[path], function)
	return
}

func (p *Project) NextEvent(cb func(config.Function)) {
	for fn := range p.notify {
		cb(fn)
	}
}

var debounceTime = time.Second * 2

func (p *Project) Start() (err error) {
	go func() {
		// every 1/4 second
		if err := p.watcher.Start(time.Millisecond * 250); err != nil {
			// watcher will only error if we already have a watcher running, so let's panic
			panic(err)
		}
	}()
	go func() {
		for event := range p.watcher.Event {
			for loc, fns := range p.functionLocations {
				if strings.HasPrefix(event.Path, loc) {
					for _, fn := range fns {
						if t := p.fnTimerMap[fn.Name]; t != nil {
							t.Stop()
						}
						fnc := fn // copy the value
						p.fnTimerMap[fn.Name] = time.AfterFunc(debounceTime, func() {
							p.notify <- fnc
						})
					}
				}
			}
		}
	}()
	return
}
