package filesystem

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"testing"

	"embly/pkg/tester"
)

func keys(m map[string]os.FileInfo) (out []string) {
	for k := range m {
		out = append(out, k)
	}
	return
}

func tarToMap(archive io.Reader) (out map[string]os.FileInfo, err error) {
	out = map[string]os.FileInfo{}

	gzr, err := gzip.NewReader(archive)
	if err != nil {
		return
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return out, err
		}
		out[hdr.Name] = hdr.FileInfo()
	}
	return
}

func TestZip(te *testing.T) {
	t := tester.New(te)
	archive, cfg, err := example.Bundle(false)
	_ = cfg
	t.Assert().NoError(err)

	files, err := tarToMap(archive)
	t.Assert().NoError(err)
	fmt.Println(keys(files))
	t.Assert().ElementsMatch(keys(files), []string{
		"/",
		"/static",
		"/static/hello",
		"/static/hello/index.html",
		"/embly_build",
		"/embly.hcl",
		"/embly_build/foo.wasm",
		"/static/index.html",
		"/data.proto",
	})

	t.Assert().True(files["/"].IsDir())
	t.Assert().True(files["/static"].IsDir())
	t.Assert().True(files["/embly_build"].IsDir())
	t.Assert().False(files["/embly.hcl"].IsDir())
	t.Assert().False(files["/embly_build/foo.wasm"].IsDir())
}
