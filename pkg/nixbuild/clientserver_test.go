package nixbuild

import (
	"embly/pkg/config"
	"embly/pkg/tester"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientServer(te *testing.T) {
	t := tester.New(te)
	if os.Getenv("NIXBUILD_INTEGRATION_TEST") == "" {
		t.Skip()
		return
	}

	cfg, err := config.New("../../examples/kv")
	if err != nil {
		return
	}

	builder, err := NewBuilder()
	t.PanicOnErr(builder.writeNixFiles())

	t.PanicOnErr(err)
	builder.SetProject(cfg)

	go func() {
		if err := builder.StartServer(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	t.PanicOnErr(builder.connectToBuildServer())
	result, err := builder.startRemoteBuild("main")
	fmt.Println(result)
	t.PanicOnErr(err)
}
