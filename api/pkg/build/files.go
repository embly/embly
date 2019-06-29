package build

import (
	"path/filepath"

	rc "embly/api/pkg/rustcompile/proto"
)

type projectFiles struct {
	files map[string]*rc.File
}

func newProjectFiles(files []*rc.File) (pf projectFiles) {
	pf.files = make(map[string]*rc.File, len(files))
	for _, f := range files {
		cleanPath := filepath.Clean(f.Path)
		f.Path = cleanPath
		pf.files[cleanPath] = f
	}
	return
}

var defaultCargoToml = []byte(`
[package]
name = ""

[dependencies]
embly="*"
`)

func (pf *projectFiles) validateAndClean() (err error) {
	foundCargotml := false
	rustFiles := 0
	for path, file := range pf.files {
		if file == nil {
			continue
		}
		ext := filepath.Ext(path)
		if path == "Cargo.toml" {
			foundCargotml = true
			if err = validateCargoToml(file.Body); err != nil {
				return
			}
			continue
		}
		if ext != ".rs" {
			delete(pf.files, path)
		}
	}
	_, _ = foundCargotml, rustFiles
	if !foundCargotml {
		pf.files["Cargo.toml"] = &rc.File{Path: "Cargo.toml", Body: defaultCargoToml}
	}
	return nil
}

func (pf *projectFiles) toCode() (code rc.Code) {
	for _, f := range pf.files {
		code.Files = append(code.Files, f)
	}
	return
}

func validateCargoToml(body []byte) (err error) {
	// TODO: disallow custom build and other unsafe features
	return nil
}
