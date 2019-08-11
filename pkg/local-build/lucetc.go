package localbuild

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os/exec"
)

var bindings = map[string]map[string]string{
	"wasi_unstable": {
		"args_get":            "__wasi_args_get",
		"args_sizes_get":      "__wasi_args_sizes_get",
		"environ_get":         "__wasi_environ_get",
		"environ_sizes_get":   "__wasi_environ_sizes_get",
		"fd_fdstat_get":       "__wasi_fd_fdstat_get",
		"fd_prestat_dir_name": "__wasi_fd_prestat_dir_name",
		"fd_prestat_get":      "__wasi_fd_prestat_get",
		"fd_write":            "__wasi_fd_write",
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

func writeBindingsFile() (location string, err error) {
	file, err := ioutil.TempFile("", "embly-bindings")
	if err != nil {
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

func runLucetc(bindingsLocation, wasmLocation, output string) (err error) {
	cmd := exec.Command("lucetc",
		"--bindings", bindingsLocation,
		"--opt-level", "2",
		"--output", output,
		wasmLocation)
	b, err := cmd.CombinedOutput()
	if err != nil {
		if len(b) > 0 {
			err = errors.New(string(b))
		}
	}
	return
}
