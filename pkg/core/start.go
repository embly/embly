package core

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"

	"embly/pkg/build"
	"embly/pkg/config"
	"embly/pkg/core/httpproto"
	"embly/pkg/dock"
	protoutil "embly/pkg/protoutil"

	vinyl "github.com/embly/vinyl/vinyl-go"
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
		descriptor, err := dock.DescriptorForFile(filepath.Join(builder.Config.ProjectRoot, db.Definition))
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
			respProto, err := httpproto.DumpRequest(r)
			if err != nil {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
			}
			if err := protoutil.WriteMessage(masterG, &respProto); err != nil {
				return err
			}
			protoRW := httpproto.ReadWriter{ReadWriter: masterG}
			if r.Body != nil {
				// async?
				_, _ = io.Copy(&protoRW, r.Body)
			}
			httpProto, err := protoRW.Next()
			if err != nil {
				return err
			}
			// defaults to 200 if we don't write it
			for k, values := range httpProto.Headers {
				for _, v := range values.Header {
					w.Header().Add(k, v)
				}
			}
			if httpProto.Status != 0 {
				w.WriteHeader(int(httpProto.Status))
			}
			_, _ = w.Write(httpProto.Body)
			for !httpProto.Eof {
				httpProto, err = protoRW.Next()
				if err != nil {
					return err
				}
				if _, err = w.Write(httpProto.Body); err != nil {
					break
				}
			}
			master.StopFunction(masterFn)
			return nil
		}()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (master *Master) makeFunctionHandler(function string) http.Handler {
	return logHandler(routeLogHandler(
		http.HandlerFunc(master.functionHandlerFunc(function)),
		master.ui,
		fmt.Sprintf("Processing by function \"%s\"", function),
	), master.ui)
}

func (master *Master) launchHTTPGateway(cfg *config.Config, g config.Gateway) (err error) {
	defaultPort := 9276
	if g.Port == 0 {
		g.Port = defaultPort
	}

	handler := http.NewServeMux()
	if g.Function != "" {
		handler.Handle("/", master.makeFunctionHandler(g.Function))
	}

	for _, route := range g.Routes {
		if route.Function != "" {
			handler.Handle(route.Path, master.makeFunctionHandler(route.Function))
		} else if route.Files != "" {
			file := cfg.GetFiles(route.Files)
			filepath := filepath.Join(
				master.builder.Config.ProjectRoot,
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
	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}

	}()
	return nil
}
