package main

// #include <stdlib.h>
//
// extern int32_t _read(void *context, int32_t id, int32_t a, int32_t b, int32_t c);
// extern int32_t _write(void *context, int32_t id, int32_t a, int32_t b, int32_t c);
// extern int32_t fd_prestat_get(void *context, int32_t a, int32_t b);
// extern int32_t fd_prestat_dir_name(void *context, int32_t a, int32_t b, int32_t c);
// extern int32_t environ_sizes_get(void *context, int32_t a, int32_t b);
// extern int32_t environ_get(void *context, int32_t a, int32_t b);
// extern int32_t args_sizes_get(void *context, int32_t a, int32_t b);
// extern int32_t args_get(void *context, int32_t a, int32_t b);
// extern int32_t fd_write(void *context, int32_t a, int32_t b, int32_t c, int32_t d);
// extern void proc_exit(void *context, int32_t a);
// extern int32_t fd_fdstat_get(void *context, int32_t a, int32_t b);
import "C"

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/pkg/errors"
	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

func main() {
	if err := run(); err != nil {
		log.Println(wasm.GetLastError())
		log.Fatalf("%+v", err)
	}
}

// wasm.Instance
func run() (err error) {
	bytes, err := wasm.ReadBytes(os.Getenv("WASM_LOCATION"))
	if err != nil {
		return errors.Wrap(err, "reading the file failed")
	}
	imports, err := createImports()
	if err != nil {
		return errors.Wrap(err, "imports failed")
	}
	instance, err := wasm.NewInstanceWithImports(bytes, imports)
	if err != nil {
		return errors.Wrap(err, "instantiation failed")
	}
	defer instance.Close()

	main := instance.Exports["main"]
	main(0, 0)

	return nil
}

type imp struct {
	name string
	fn   interface{}
	ptr  unsafe.Pointer
}

var importMap = map[string][]imp{
	"wasi_unstable": []imp{
		{name: "fd_write", fn: fd_write, ptr: C.fd_write},
		{name: "fd_write", fn: fd_write, ptr: C.fd_write},
		{name: "fd_prestat_get", fn: fd_prestat_get, ptr: C.fd_prestat_get},
		{name: "fd_prestat_dir_name", fn: fd_prestat_dir_name, ptr: C.fd_prestat_dir_name},
		{name: "environ_sizes_get", fn: environ_sizes_get, ptr: C.environ_sizes_get},
		{name: "environ_get", fn: environ_get, ptr: C.environ_get},
		{name: "args_sizes_get", fn: args_sizes_get, ptr: C.args_sizes_get},
		{name: "args_get", fn: args_get, ptr: C.args_get},
		{name: "proc_exit", fn: proc_exit, ptr: C.proc_exit},
		{name: "fd_fdstat_get", fn: fd_fdstat_get, ptr: C.fd_fdstat_get},
	},
	"env": []imp{
		{name: "_read", fn: _read, ptr: C._read},
		{name: "_write", fn: _write, ptr: C._write},
	},
}

func createImports() (imports *wasm.Imports, err error) {
	imports = wasm.NewImports()
	for module, imps := range importMap {
		imports.Namespace(module)
		for _, i := range imps {
			if imports, err = imports.Append(i.name, i.fn, i.ptr); err != nil {
				return
			}
		}
	}
	return
}

func getUint32(addr int, mem *wasm.Memory) uint32 {
	return binary.LittleEndian.Uint32(mem.Data()[addr:])
}

func setUint32(addr int, v uint32, mem *wasm.Memory) {
	binary.LittleEndian.PutUint32(mem.Data()[addr:], v)
}

func setInt32(addr int, v int32, mem *wasm.Memory) {
	setUint32(addr, uint32(v), mem)
}

//export _read
func _read(context unsafe.Pointer, id int32, a int32, b int32, writtenPtr int32) int32 {
	ctx := wasm.IntoInstanceContext(context)
	_ = ctx
	fmt.Println("_read", id, a, b, writtenPtr)
	return 0
}

//export _write
func _write(context unsafe.Pointer, id int32, buffPtr int32, buffLen int32, writtenPtr int32) int32 {
	ctx := wasm.IntoInstanceContext(context)
	buff := int(getUint32(int(buffPtr), ctx.Memory()))
	fmt.Println("_write", id, buffPtr, buff, buffLen, writtenPtr)
	os.Stdout.Write(ctx.Memory().Data()[int(buffPtr) : int(buffPtr)+int(buffLen)])
	setInt32(int(writtenPtr), buffLen, ctx.Memory())
	return 0
}

//export fd_write
func fd_write(context unsafe.Pointer, fd int32, addr int32, ln int32, written int32) int32 {
	ctx := wasm.IntoInstanceContext(context)
	add := int(getUint32(int(addr), ctx.Memory()))
	l := int(ln)
	os.Stdout.Write(ctx.Memory().Data()[add : add+l])
	setInt32(int(written), ln, ctx.Memory())
	return 0
}

//export fd_prestat_get
func fd_prestat_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("fd_prestat_get", a, b)
	return 0
}

//export fd_prestat_dir_name
func fd_prestat_dir_name(context unsafe.Pointer, a int32, b int32, c int32) int32 {
	fmt.Println("fd_prestat_dir_name")
	return 0
}

//export environ_sizes_get
func environ_sizes_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("environ_sizes_get")
	return 0
}

//export environ_get
func environ_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("environ_get")
	return 0
}

//export args_sizes_get
func args_sizes_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("args_sizes_get")
	return 0
}

//export args_get
func args_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("args_get")
	return 0
}

//export proc_exit
func proc_exit(context unsafe.Pointer, a int32) {
	fmt.Println("proc_exit")
}

//export fd_fdstat_get
func fd_fdstat_get(context unsafe.Pointer, a int32, b int32) int32 {
	fmt.Println("fd_fdstat_get")
	return 0
}
