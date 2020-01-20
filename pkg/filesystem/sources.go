package filesystem

import (
	"io"
	"path/filepath"
	"strings"

	"embly/pkg/config"

	"github.com/pkg/errors"
	"golang.org/x/tools/godoc/vfs"
)

func ZipSources(projectRoot string, buildLocation string, sources []string) (
	newBuildLocation string, archive io.Reader, err error) {
	g := config.Gateway{}
	_ = g
	var locations []string
	for _, s := range sources {
		locations = append(locations, filepath.Join(projectRoot, s))
	}

	buildLocation = filepath.Join(projectRoot, buildLocation)
	locations = append(locations, buildLocation)
	buildRoot := CommonPrefix(locations)

	ns := vfs.NameSpace{}

	newLocation := func(old string) string {
		return "/" + strings.TrimPrefix(old, buildRoot)
	}

	for _, l := range locations {
		newLoc := newLocation(l)
		ns.Bind(newLoc, vfs.OS(l), "/", vfs.BindAfter)
	}

	fs := FileSystem{ns}

	a, err := fs.zip([]string{"./"}, nil, nil)
	if err != nil {
		err = errors.WithStack(err)
	}

	return strings.TrimPrefix(buildLocation, buildRoot), a, err
}
