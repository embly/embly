package filesystem

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// walk recursively descends path, calling walkFn.
func (fs *FileSystem) walk(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := fs.readDirNames(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := fs.lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = fs.walk(filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func (fs *FileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	info, err := fs.lstat(root)

	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = fs.walk(root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// lstat is our specific lstat for walk. Sometimes a directory structure might
// exist, but the actual directly is not present. We still want to walk the directory
// so we have to move out the relevant fileinfo
func (fs *FileSystem) lstat(path string) (fi os.FileInfo, err error) {
	info, err := fs.Lstat(path)
	if err == nil {
		return info, err
	}
	name := filepath.Base(path)
	if name == "." {
		name = "/"
	}
	return dirInfo(name), nil
}

// dirInfo is a trivial implementation of os.FileInfo for a directory.
type dirInfo string

func (d dirInfo) Name() string       { return string(d) }
func (d dirInfo) Size() int64        { return 0 }
func (d dirInfo) Mode() os.FileMode  { return os.ModeDir | 0555 }
func (d dirInfo) ModTime() time.Time { return startTime }
func (d dirInfo) IsDir() bool        { return true }
func (d dirInfo) Sys() interface{}   { return nil }

var startTime = time.Now()

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func (fs *FileSystem) readDirNames(dirname string) ([]string, error) {
	files, err := fs.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	sort.Strings(names)
	return names, nil
}
