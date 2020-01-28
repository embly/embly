package filesystem

import (
	"crypto/sha256"
	"embly/pkg/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
)

type FileInfo struct {
	os.FileInfo
	Hash []byte
}

func (fi *FileInfo) PopulateHash(path string) (err error) {
	if fi.IsDir() {
		return
	}
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		err = errors.WithStack(err)
		return
	}
	fi.Hash = h.Sum(nil)
	return
}

type Project struct {
	watcher           *watcher.Watcher
	cfg               *config.Config
	functionLocations map[string][]config.Function

	notify chan config.Function

	fnMap      map[string]config.Function
	fnTimerMap map[string]*time.Timer

	files map[string]*FileInfo
}

// NewProject should create a new project
func NewProject(cfg *config.Config) (p *Project) {
	p = &Project{
		watcher:           watcher.New(),
		cfg:               cfg,
		notify:            make(chan config.Function, 100),
		fnMap:             map[string]config.Function{},
		fnTimerMap:        map[string]*time.Timer{},
		functionLocations: map[string][]config.Function{},
		files:             map[string]*FileInfo{},
	}
	for _, fn := range cfg.Functions {
		p.fnMap[fn.Name] = fn
	}
	return
}

func (p *Project) FunctionWatchedFiles(name string) (path string, files map[string]*FileInfo, err error) {
	// quick and easy hack, just use watcher to crawl the files
	tmpP := NewProject(p.cfg)
	selected, ok := tmpP.fnMap[name]
	if !ok {
		err = errors.Errorf(`couldn't find function with name "%s"`, name)
		return
	}
	path = filepath.Join(p.cfg.ProjectRoot, selected.Path) + "/"
	if err = tmpP.AddFunctionFiles(selected); err != nil {
		err = errors.WithStack(err)
		return
	}
	filesInt := tmpP.watcher.WatchedFiles()
	files = make(map[string]*FileInfo)
	for name, fi := range filesInt {
		files[name] = &FileInfo{FileInfo: fi}
	}
	return
}

func (p *Project) FunctionSources(name string) (path string, files []string, err error) {
	path, fileMap, err := p.FunctionWatchedFiles(name)
	for name := range fileMap {
		files = append(files, name)
	}
	return
}

func (p *Project) CopyFunctionSourcesToBuild(path, name string) (buildDir string, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	functionPath, files, err := p.FunctionSources(name)
	if err != nil {
		return
	}
	prefix := CommonPrefix(files)
	sort.Strings(files)
	buildDir, err = ioutil.TempDir(path, "embly-build")
	if err != nil {
		return
	}
	if err = os.Chmod(buildDir, os.ModePerm); err != nil {
		return
	}

	for _, file := range files {
		var fi os.FileInfo
		fi, err = os.Stat(file)
		if err != nil {
			return
		}
		newLoc := filepath.Join(buildDir, strings.TrimPrefix(file, prefix))
		if fi.IsDir() {
			if err = os.MkdirAll(newLoc, os.ModeDir|os.ModePerm); err != nil {
				return
			}
		} else {
			if err = CopyFile(file, newLoc); err != nil {
				return
			}
		}
	}

	// don't return the base buildDir, return the path where the function is
	buildDir = filepath.Join(buildDir, strings.TrimPrefix(functionPath, prefix))

	return
}

func CopyFile(src, dest string) (err error) {
	from, err := os.Open(src)
	if err != nil {
		return
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return
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

func (p *Project) AddFunctionFiles(fn config.Function) (err error) {
	if err = p.AddRecursive(fn.Path, fn); err != nil {
		return
	}
	for _, source := range fn.Sources {
		if err = p.AddRecursive(source, fn); err != nil {
			return
		}
	}
	return
}
func (p *Project) Start() (err error) {
	// if functions depend on shared sources watcher.Watcher will de-dupe them
	for _, fn := range p.cfg.Functions {
		if err = p.AddFunctionFiles(fn); err != nil {
			return
		}
	}

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
