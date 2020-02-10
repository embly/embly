package lucet

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var bindings = map[string]map[string]string{
	"wasi_unstable": {
		"args_get":            "__wasi_args_get",
		"args_sizes_get":      "__wasi_args_sizes_get",
		"clock_res_get":       "__wasi_clock_res_get",
		"clock_time_get":      "__wasi_clock_time_get",
		"environ_get":         "__wasi_environ_get",
		"environ_sizes_get":   "__wasi_environ_sizes_get",
		"fd_fdstat_get":       "__wasi_fd_fdstat_get",
		"fd_prestat_dir_name": "__wasi_fd_prestat_dir_name",
		"fd_prestat_get":      "__wasi_fd_prestat_get",
		"fd_write":            "__wasi_fd_write",
		"poll_oneoff":         "__wasi_poll_oneoff",
		"proc_exit":           "__wasi_proc_exit",
		"random_get":          "__wasi_random_get",
	},
	"wasi_snapshot_preview1": {
		"args_get":            "__wasi_args_get",
		"args_sizes_get":      "__wasi_args_sizes_get",
		"clock_res_get":       "__wasi_clock_res_get",
		"clock_time_get":      "__wasi_clock_time_get",
		"environ_get":         "__wasi_environ_get",
		"environ_sizes_get":   "__wasi_environ_sizes_get",
		"fd_fdstat_get":       "__wasi_fd_fdstat_get",
		"fd_prestat_dir_name": "__wasi_fd_prestat_dir_name",
		"fd_prestat_get":      "__wasi_fd_prestat_get",
		"fd_write":            "__wasi_fd_write",
		"poll_oneoff":         "__wasi_poll_oneoff",
		"proc_exit":           "__wasi_proc_exit",
		"random_get":          "__wasi_random_get",
	},
	"embly": {
		"_read":   "__read",
		"_write":  "__write",
		"_spawn":  "__spawn",
		"_events": "__events",
	},
}

// WriteBindingsFile will write lucet binding to a file
func WriteBindingsFile() (location string, err error) {
	file, err := ioutil.TempFile("", "embly-bindings")
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	b, err := json.Marshal(bindings)
	if err != nil {
		return
	}

	if _, err = file.Write(b); err != nil {
		return
	}

	location = file.Name()
	return
}

// RunLucetc run lucet locally
func RunLucetc(bindingsLocation, wasmLocation, out string) (err error) {
	// TODO? seems to result in about a 10% size savings
	// also: https://github.com/fastly/lucet/issues/208#issue-455070676
	// cmd := exec.Command("wasm-opt",
	// 	"-o", wasmLocation,
	// 	"-Os",
	// 	wasmLocation)
	// b, err := cmd.CombinedOutput()
	// if err != nil {	"embly/pkg/randy"

	// }
	cmd := exec.Command("lucetc",
		"--bindings", bindingsLocation,
		// "--opt-level", "2",
		"--opt-level", "0",
		"--emit", "so",
		"--output", out,
		wasmLocation)
	b, err := cmd.CombinedOutput()
	if err != nil {
		if len(b) > 0 {
			err = errors.New(string(b))
		}
	}
	return
}

func hashFile(file string) (out string, err error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	sum := sha256.Sum256(b)
	out = fmt.Sprintf("%x", sum)
	return
}

func cacheDir() (dir string, err error) {
	homedir, err := homedir.Dir()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	dir = filepath.Join(homedir, "./.embly/lucet_cache")
	return
}

func createHomeDir() (err error) {
	dir, err := cacheDir()
	_, err = os.Stat(dir)
	if err != nil {
		fmt.Println("Creating ~/.embly directory to store cached values")
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			err = errors.WithStack(err)
			return err
		}
	}
	return nil
}

// CompileWasmToObject runs lucetc with our bindings to create a local object file
func CompileWasmToObject(file, out string) (err error) {
	if err = createHomeDir(); err != nil {
		return
	}

	hash, err := hashFile(file)
	if err != nil {
		return err
	}

	dir, _ := cacheDir()
	hashDestination := filepath.Join(dir, hash)
	_, err = os.Stat(hashDestination)
	if err != nil {
		bindingsLocation, err := WriteBindingsFile()
		defer os.Remove(bindingsLocation)
		if err != nil {
			return err
		}
		err = RunLucetc(bindingsLocation, file, hashDestination)
		if err != nil {
			return err
		}
	}

	b, err := ioutil.ReadFile(hashDestination)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(out, b, 0644)
}
