package rustcompile

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	rc "embly/pkg/rustcompile/proto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Start ...
func Start(port int) {
	flag.Parse()
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc.WithInsecure()
	grpcServer := grpc.NewServer()
	rc.RegisterRustCompileServer(grpcServer, &server{})
	logrus.Info("Serving rustcompile on " + addr)
	grpcServer.Serve(lis)
}

type server struct{}

func (s *server) StartBuild(code *rc.Code, stream rc.RustCompile_StartBuildServer) (err error) {
	var tmpdir string
	if tmpdir, err = ioutil.TempDir("", "rustcompile"); err != nil {
		return
	}
	var resultTmpdir string
	if resultTmpdir, err = ioutil.TempDir("", "rustcompile"); err != nil {
		return
	}

	for _, file := range code.Files {
		location := path.Join(tmpdir, file.Path)
		directory := filepath.Dir(location)
		if err = os.MkdirAll(directory, os.ModePerm); err != nil {
			return
		}
		if err = ioutil.WriteFile(path.Join(tmpdir, file.Path), file.Body, 0644); err != nil {
			return
		}
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf(`cd %s \
	&& cargo +nightly build --offline --target wasm32-wasi --release -Z unstable-options --out-dir %s \
	&& wasm-strip %s/*.wasm \
	&& ls -lah %s/*.wasm
`, tmpdir, resultTmpdir, resultTmpdir, resultTmpdir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return err
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	fmt.Println("command finished")
	return nil
}

func indexHandler(ctx context.Context, db *sql.DB, c *gin.Context) error {
	c.JSON(200, gin.H{"msg": "Hello"})
	return nil
}
