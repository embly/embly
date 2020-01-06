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

	"net/http"
	"net/http/httputil"
	"net/url"

	vinyl "embly/pkg/vinyl"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

type StartConfig struct {
	Watch bool
	Dev   bool
	Host  string
}

// Start starts
func Start(builder *build.Builder, ui cli.Ui, startConfig StartConfig) (err error) {
	ui.Info("Starting dev server")
	master := NewMaster()
	master.ui = ui
	master.host = startConfig.Host
	master.builder = builder
	master.developmentRun = startConfig.Dev
	master.databases = make(map[string]config.Database)
	for name, fn := range builder.Functions {
		master.RegisterFunctionName(name, fn.Obj)
		ui.Output(fmt.Sprintf("Registering %s with %s", name, fn.Obj))
	}

	for _, db := range builder.Config.Databases {
		ui.Info(fmt.Sprintf("Configuring database \"%s\"", db.Name))
		v, err := dock.StartVinyl(db.Name)
		if err != nil {
			return err
		}

		ui.Output("Parsing file descriptor")
		descriptor, err := dock.DescriptorForFile(filepath.Join(builder.ProjectRoot, db.Definition))
		if err != nil {
			return err
		}
		// TODO: handle this in vinyl-go
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(descriptor); err != nil {
			return err
		}
		zw.Close()

		metadata := db.ToMetadata()
		metadata.Descriptor = buf.Bytes()

		ui.Output("Waiting for database to start")
		if err := v.Wait(); err != nil {
			return err
		}

		db.DB, err = vinyl.Connect(
			fmt.Sprintf("vinyl://what:ever@localhost:%d/foo", v.Port), metadata)
		if err != nil {
			return err
		}
		ui.Output("Connected to database")
		db.Port = v.Port
		master.databases[db.Name] = db
	}

	go master.Start()
	if startConfig.Watch {
		ui.Info("Watching for local changes")
		if err := builder.WatchForChangesAndRebuild(); err != nil {
			return errors.Wrap(err, "error watching for changes")
		}
	}
	waitChan := make(chan struct{})
	for _, g := range builder.Config.Gateways {
		switch kind := g.Type; kind {
		case "http":
			if err := master.launchHTTPGateway(builder.Config, g); err != nil {
				return err
			}
		default:
			return errors.Errorf("gateway type of '%s' not available", kind)
		}
	}
	<-waitChan
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
			if r.Body != nil {
				// async?
				io.Copy(masterG, r.Body)
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
				logHandler(routeLogHandler(
					http.HandlerFunc(master.functionHandlerFunc(route.Function)),
					master.ui,
					fmt.Sprintf("Processing by function \"%s\"", route.Function),
				), master.ui),
			)
		} else if route.Files != "" {
			file := cfg.GetFiles(route.Files)
			filepath := filepath.Join(
				master.builder.ProjectRoot,
				file.Path)
			var h http.Handler

			h = http.FileServer(http.Dir(
				filepath,
			))

			if file.LocalFileServer != "" && master.developmentRun {
				u, err := url.Parse(file.LocalFileServer)
				if err != nil {
					// TODO: handle
					panic(err)
				}
				h = httputil.NewSingleHostReverseProxy(u)
			}
			master.ui.Info(fmt.Sprintf("Registering static files at path %s", route.Path))
			handler.Handle(route.Path,
				singleLineLogHandler(
					http.StripPrefix(route.Path, h),
					master.ui,
					fmt.Sprintf("[%s]: ", route.Files),
				),
			)
		}
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", master.host, g.Port),
		Handler: handler,
	}
	master.ui.Info(fmt.Sprintf("HTTP gateway listening on port %d\n", g.Port))
	go server.ListenAndServe()
	return nil
}
