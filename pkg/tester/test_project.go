package tester

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateTmpTestProject(sources ...map[string]string) (projectRoot string, err error) {

	dir, err := ioutil.TempDir("", "embly-test")
	if err != nil {
		return
	}

	// create a subdirectory so that we can test deps outside of the project
	projectRoot = filepath.Join(dir, "./project")

	if err = os.Mkdir(projectRoot, os.ModePerm); err != nil {
		return
	}

	if len(sources) == 0 {
		sources = []map[string]string{
			ExternalDependency,
			ExternalDependency2,
			EmblyFile,
			BasicRustProject,
			FooRustProject,
		}
	}

	for _, source := range sources {
		for name, body := range source {
			loc := filepath.Join(projectRoot, name)
			if err = os.MkdirAll(filepath.Dir(loc), os.ModePerm); err != nil {
				return
			}
			if err = ioutil.WriteFile(loc, []byte(body), 0644); err != nil {
				return
			}
		}
	}
	return
}
