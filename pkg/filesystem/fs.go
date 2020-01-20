package filesystem

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"embly/pkg/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

type FileSystem struct {
	vfs.FileSystem
}

// you must use relative paths without dot (./) notation
var exampleFiles = map[string]string{
	"embly.hcl": `
function "foo" {
	runtime = "rust"
	path = "./foo"
}
files "static" {
	path = "./static"
}
database "vinyl" "main" {
	definition = "data.proto"
}
`,
	"Cargo.lock":              "asdfadsf",
	"embly_build/foo.wasm":    "contents",
	"embly_build/foo.out":     "contents",
	"data.proto":              "o",
	"project/Cargo.lock":      "locked",
	"project/src/main.rs":     "println(\"hi\");",
	"static/index.html":       "<body></body>",
	"static/hello/index.html": "<body></body>",
}
var example = FileSystem{mapfs.New(exampleFiles)}

func (fs *FileSystem) zip(directories, files, globs []string) (archive io.Reader, err error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	filesAndDirsToAdd := map[string]struct{}{
		"/": struct{}{},
	}

	addFile := func(path string, info os.FileInfo, err error) error {
		// TODO: this is just artbitrary and hardcoded, fix that
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "target" {
			return filepath.SkipDir
		}
		filesAndDirsToAdd[path] = struct{}{}
		return nil
	}

	var glob string
	addGlob := func(path string, info os.FileInfo, err error) error {
		var matched bool
		if info.IsDir() && strings.HasPrefix(glob, path) {
			return addFile(path, info, err)
		}
		matched, err = filepath.Match(glob, path)
		if err != nil {
			return err
		}
		if matched {
			return addFile(path, info, err)
		}
		return nil
	}

	for _, directory := range directories {
		if err = fs.Walk(directory, addFile); err != nil {
			return
		}
	}
	var fi os.FileInfo
	for _, file := range files {
		fi, err = fs.Stat(file)
		if err != nil {
			return
		}
		if err = addFile(file, fi, nil); err != nil {
			return
		}
	}
	for _, glob = range globs {
		if err = fs.Walk("/", addGlob); err != nil {
			return
		}
	}

	// we need to sort so that directories are unpacked before
	// their files
	sorted := []string{}
	for path := range filesAndDirsToAdd {
		sorted = append(sorted, path)
	}
	sort.Strings(sorted)

	for _, path := range sorted {
		var info os.FileInfo
		info, err = fs.lstat(path)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
		var hdr *tar.Header
		hdr, err = tar.FileInfoHeader(info, "")
		if err != nil {
			return
		}
		hdr.Name = path

		var b []byte
		if !info.IsDir() {
			var f io.Reader
			f, err = fs.Open(path)
			if err != nil {
				return
			}
			b, err = ioutil.ReadAll(f)
			if err != nil {
				return
			}
			hdr.Size = int64(len(b))
		}

		if err = tw.WriteHeader(hdr); err != nil {
			return
		}

		if _, err = tw.Write(b); err != nil {
			return
		}
	}
	return &buf, nil
}

func (fs *FileSystem) Bundle(includeObjectFiles bool) (archive io.Reader, cfg *config.Config, err error) {
	cfgFile, err := fs.Open("embly.hcl")
	if err != nil {
		return
	}

	cfg, err = config.ParseConfig(cfgFile)
	if err != nil {
		return
	}

	directories := []string{}
	for _, f := range cfg.Files {
		directories = append(directories, filepath.Join("/", f.Path))
	}

	files := []string{"/embly.hcl"}
	for _, db := range cfg.Databases {
		files = append(files, filepath.Join("/", db.Definition))
	}
	globs := []string{"/embly_build/*.wasm"}
	if includeObjectFiles {
		globs = append(globs, "/embly_build/*."+runtime.GOOS)
	}
	archive, err = fs.zip(directories, files, globs)
	return archive, cfg, err
}
