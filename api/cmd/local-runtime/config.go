package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type config struct {
	verbose          bool
	debug            bool
	emblyProjectFile string
}

func flags() {
	cfg := config{}
	flag.BoolVar(&cfg.verbose, "v", false, "enable verbose logging")
	flag.BoolVar(&cfg.debug, "d", false, "print stdout and stderr from wasm")
	// flag.BoolVar(&cfg.emblyProjectFile, "f", false, "")
	flag.Parse()
	c = &cfg
}

var c *config

type emblyProject struct {
	Functions map[string]projectFunction `json:"functions"`
}

type projectFunction struct {
	Path    string `json:"path"`
	Runtime string `json:"runtime"`
}

func getEmblyProjectFile() (ep *emblyProject, err error) {
	var f *os.File
	if f, err = findConfigFile(); err != nil {
		return
	}
	if ep, err = parseConfigFile(f); err != nil {
		return
	}
	err = validateEmblyProjectFile(ep)
	return
}

func parseConfigFile(f *os.File) (ep *emblyProject, err error) {
	var b []byte
	if b, err = ioutil.ReadAll(f); err != nil {
		return
	}
	ep = &emblyProject{}
	err = yaml.Unmarshal(b, ep)
	return
}

func validateEmblyProjectFile(ep *emblyProject) (err error) {
	if ep.Functions == nil {
		return errors.New("no functions in embly-project.yml file")
	}
	for name, spec := range ep.Functions {
		// todo: validate name?
		if spec.Path == "" {
			return errors.Errorf("function %s must have a path", name)
		}
		if spec.Runtime == "" {
			return errors.Errorf("function %s must have a runtime value", name)
		}
	}
	return
}

func findConfigFile() (f *os.File, err error) {
	var wd string
	if wd, err = os.Getwd(); err != nil {
		return
	}
	for {
		if f, err = os.Open(filepath.Join(wd, "./embly-project.yml")); err == nil {
			break
		}
		parent := filepath.Join(wd, "../")
		if wd == parent || wd == "/" {
			break
		}
		wd = parent
	}
	if f == nil {
		err = errors.New("embly-project.yml not found in this directory or any parent")
		return
	}
	return
}
