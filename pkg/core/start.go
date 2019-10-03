package core

import (
	"bytes"
	"compress/gzip"
	"embly/pkg/build"
	"embly/pkg/config"
	"embly/pkg/dock"
	"embly/pkg/proxy"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"net/http"

	vinyl "embly/pkg/vinyl"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

type vinylContainer struct {
	db        *vinyl.DB
	container *dock.Vinyl
}

// Start starts
func Start(co build.CompileOutput, ui cli.Ui) (err error) {
	ui.Info("Starting dev server")
	master := NewMaster()
	master.ui = ui
	master.compileOutput = co
	master.databases = make(map[string]config.Database)
	for name, fn := range co.Functions {
		master.RegisterFunctionName(name, fn.Obj)
		ui.Output(fmt.Sprintf("Registering %s with %s", name, fn.Obj))
	}

	for _, db := range co.Config.Databases {
		ui.Info(fmt.Sprintf("Configuring database \"%s\"", db.Name))
		v, err := dock.StartVinyl(db.Name)
		if err != nil {
			return err
		}

		ui.Output("Parsing file descriptor")
		descriptor, err := dock.DescriptorForFile(filepath.Join(co.ProjectRoot, db.Definition))
		if err != nil {
			return err
		}
		// TODO: handle this in vinyl-go
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		zw.Write(descriptor)
		zw.Close()

		metadata := db.ToMetadata()
		metadata.Descriptor = buf.Bytes()

		if err := v.Wait(); err != nil {
			return err
		}

		db.DB, err = vinyl.Connect(
			fmt.Sprintf("vinyl://what:ever@localhost:%d/foo", v.Port), metadata)
		if err != nil {
			return err
		}
		db.Container = v
		master.databases[db.Name] = db
	}

	go master.Start()
	var wg sync.WaitGroup
	wg.Add(1)
	for _, g := range co.Config.Gateways {
		switch kind := g.Type; kind {
		case "http":
			if err := master.launchHTTPGateway(co.Config, g); err != nil {
				return err
			}
		default:
			return errors.Errorf("gateway type of '%s' not available", kind)
		}
	}
	wg.Wait()
	return nil
}

type statusWriter struct {
	writer    io.Writer
	hasHeader bool
	header    []byte

	proto      string
	status     string
	statusCode int
}

func (sw *statusWriter) Write(buf []byte) (ln int, err error) {
	if !sw.hasHeader {
		index := bytes.IndexByte(buf, byte('\n'))
		if index == -1 {
			index = len(buf) - 1
		} else {
			sw.hasHeader = true
		}
		sw.header = append(sw.header, buf[0:index]...)
	}
	return sw.writer.Write(buf)
}

type badHeaderError struct {
	message string
	line    string
}

func (bhe *badHeaderError) Error() string {
	return fmt.Sprintf("%s '%s'", bhe.message, bhe.line)
}

func (sw *statusWriter) parseHeader() (err error) {
	line := string(sw.header)
	var i int
	if i = strings.IndexByte(line, ' '); i == -1 {
		return &badHeaderError{"malformed HTTP response", line}
	}
	sw.proto = line[:i]
	sw.status = strings.TrimLeft(line[i+1:], " ")

	statusCode := sw.status
	if i := strings.IndexByte(sw.status, ' '); i != -1 {
		statusCode = sw.status[:i]
	}
	if len(statusCode) != 3 {
		return &badHeaderError{"malformed HTTP status code", statusCode}
	}
	sw.statusCode, err = strconv.Atoi(statusCode)
	if err != nil || sw.statusCode < 0 {
		return &badHeaderError{"malformed HTTP status code", statusCode}
	}
	// todo: proto
	return nil
}

func (master *Master) functionHandlerFunc(name string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			masterG := master.NewGateway()
			masterFn, err := master.NewFunction(
				name, masterG.ID, nil, nil)
			if err != nil {
				return err
			}
			masterG.AttachFn(masterFn)
			if err := masterFn.Start(); err != nil {
				return err
			}
			out, err := proxy.DumpRequest(r)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
			}
			if _, err := masterG.Write(out); err != nil {
				return err
			}

			hj, ok := w.(http.Hijacker)
			if !ok {
				return errors.New("webserver doesn't support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				return err
			}

			// After hijack returned errors are not written as errors
			sw := &statusWriter{
				writer: conn,
			}
			_, err = io.Copy(sw, masterG)
			if err != nil {
				log.Fatal(err)
			}
			if err := sw.parseHeader(); err != nil {
				log.Fatal(err)
			}
			if rl, ok := w.(commonLoggingResponseWriter); ok {
				rl.SetStatus(sw.statusCode)
			} else {
				log.Fatal("unable to get it done")
			}
			master.StopFunction(masterFn)
			conn.Close()
			return nil
		}()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (master *Master) launchHTTPGateway(cfg config.Config, g config.Gateway) (err error) {
	defaultPort := 9276
	if g.Port == 0 {
		g.Port = defaultPort
	}

	handler := http.NewServeMux()
	if g.Function != "" {
		handler.HandleFunc("/", master.functionHandlerFunc(g.Function))
	}

	for _, route := range g.Routes {
		if route.Function != "" {
			handler.Handle(route.Path,
				routeLogHandler(
					http.HandlerFunc(master.functionHandlerFunc(route.Function)),
					master.ui,
					fmt.Sprintf("Processing by function \"%s\"", route.Function),
				),
			)
		}
		if route.Files != "" {
			filepath := filepath.Join(
				master.compileOutput.ProjectRoot,
				cfg.GetFiles(route.Files).Path)

			handler.Handle(route.Path,
				routeLogHandler(
					http.StripPrefix(route.Path,
						http.FileServer(
							http.Dir(
								filepath,
							),
						),
					),
					master.ui,
					fmt.Sprintf("Serving assets \"%s\"", route.Files),
				),
			)
		}
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", g.Port),
		Handler: logHandler(handler, master.ui),
	}
	master.ui.Info(fmt.Sprintf("HTTP gateway listening on port %d\n", g.Port))
	go server.ListenAndServe()
	return nil
}
