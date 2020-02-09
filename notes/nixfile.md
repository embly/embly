Nixfile

runs commands in nix-shell in a running docker container. syntax is with starlark. you define commands to run, they are run as a non-priveledged user. You can define inputs and outputs. The inputs can be output files from different build steps or they can be nix packages. Like so:

```python


go_download = run(
    srcs=["go.mod", "go.sum"],
    paths=["/home/"]
    buildInputs=["go"],
    environment={"GOPATH":"/home/"},
    script="""
    go mod download
"""
)

go_build = run(
    srcs=["./pkg/", "./cmd/embly", "go.mod", "go.sum"],
    paths=[f"go_download/home/"]
    buildInputs=["go"],
    environment={"GOPATH":"/home/go/"},
    script="""
    cd cmd/embly
    go build
"""
)

cmd(
    paths=[f"{go_build}/project/embly/cmd"]
    dependencies=["python3"],
    environment={"GOPATH":"/home/"},
    script="""
    /project/embly/cmd
"""
)

```

they could run at any point, if their input file change
maybe write a more complicated example of them building something
we could have the build list its output files and then only those are copied into
the final image.


blarg
