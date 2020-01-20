package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	comms_proto "embly/pkg/core/proto"
	vinyl "github.com/embly/vinyl/vinyl-go"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

// Config represents an embly.hcl config
type Config struct {
	ProjectRoot  string
	Gateways     []Gateway  `hcl:"gateway,block"`
	Dependencies []string   `hcl:"dependencies,optional"`
	Functions    []Function `hcl:"function,block"`
	Files        []Files    `hcl:"files,block"`
	filesMap     map[string]Files
	Databases    []Database `hcl:"database,block"`
	databaseMap  map[string]*Database
}

func (cfg *Config) AbsolutePath(relativePath string) string {
	return filepath.Join(cfg.ProjectRoot, relativePath)
}

func New(path string) (cfg *Config, err error) {
	f, l, err := FindConfigFile(path)
	if err != nil {
		return
	}
	cfg, err = ParseConfig(f)
	if err != nil {
		return
	}
	cfg.ProjectRoot = l
	return
}

// GetFiles retireve a "files" configuration value using a reference, like "files.foo"
func (cfg *Config) GetFiles(name string) Files {
	parts := strings.Split(name, ".")
	return cfg.filesMap[parts[1]]
}

// GetDatabase get it
func (cfg *Config) GetDatabase(name string) *Database {
	return cfg.databaseMap[name]
}

// GetProtoDBs gets the proto structure of a db definition
func (cfg *Config) GetProtoDBs() []*comms_proto.DB {
	out := make([]*comms_proto.DB, len(cfg.Databases))
	for i, db := range cfg.Databases {
		out[i] = &comms_proto.DB{
			Type:       db.Type,
			Name:       db.Name,
			Connection: db.ConnectionString(),
			Token:      db.DB.Token,
		}
	}
	return out
}

// Gateway describes public routing. An http listener, tcp, ssh, etc...
type Gateway struct {
	Type     string         `hcl:"type"`
	Port     int            `hcl:"port,optional"`
	Function string         `hcl:"function,optional"`
	Routes   []GatewayRoute `hcl:"route,block"`
}

// GatewayRoute is a specific routing rule for a gateway. Only used with the http gateway
type GatewayRoute struct {
	Function string `hcl:"function,optional"`
	Path     string `hcl:"path,label"`
	Files    string `hcl:"files,optional"`
}

// Function is an embly function, the main compute primitive in Embly
type Function struct {
	Name    string   `hcl:"name,label"`
	Runtime string   `hcl:"runtime,attr"`
	Path    string   `hcl:"path,attr"`
	Sources []string `hcl:"sources,optional"`
}

// Files are local static assets that are served by the runtime
type Files struct {
	Name            string `hcl:"name,label"`
	Path            string `hcl:"path,attr"`
	LocalFileServer string `hcl:"local_file_server,optional"`
}

// Database describes the schemas and configuration of a datastore
type Database struct {
	Type       string           `hcl:"type,label"`
	Name       string           `hcl:"name,label"`
	Definition string           `hcl:"definition,attr"`
	Records    []DatabaseRecord `hcl:"record,block"`
	DB         *vinyl.DB
	Port       int
}

// ToMetadata converts a database configuration value to vinyl Metadata. Does not populate the protobuf
// descriptor
func (db *Database) ToMetadata() (md vinyl.Metadata) {
	for _, rec := range db.Records {
		indexes := []vinyl.Index{}
		for _, idx := range rec.Indexes {
			indexes = append(indexes, vinyl.Index{
				Field:  idx.Name,
				Unique: idx.Unique,
			})
		}
		md.Records = append(md.Records, vinyl.Record{
			Name:       rec.Name,
			PrimaryKey: rec.PrimaryKey,
			Indexes:    indexes,
		})
	}
	return
}

// ConnectionString the connection string for this database type
func (db *Database) ConnectionString() string {
	return fmt.Sprintf("vinyl://user:pass@localhost:%d", db.Port)
}

// DatabaseRecord describes a database record for a vinly database
type DatabaseRecord struct {
	Name       string                `hcl:"name,label"`
	PrimaryKey string                `hcl:"primary_key,attr"`
	Indexes    []DatabaseRecordIndex `hcl:"index,block"`
}

// DatabaseRecordIndex is an index definition for a database record
type DatabaseRecordIndex struct {
	Name   string `hcl:"name,label"`
	Unique bool   `hcl:"unique,optional"`
}

// ParseConfig will parse an embly.hcl file
func ParseConfig(configFile io.Reader) (cfg *Config, err error) {
	cfg = &Config{}

	evalContext := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
	}
	_ = evalContext
	p := hclparse.NewParser()
	b, err := ioutil.ReadAll(configFile)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	file, diagnostics := p.ParseHCL(b, "embly.hcl")
	if diagnostics.HasErrors() {
		err = diagnostics
		return
	}
	_ = gohcl.DecodeBody(file.Body, nil, cfg)

	functionMap := map[string]cty.Value{}
	for _, fn := range cfg.Functions {
		functionMap[fn.Name] = cty.StringVal("function." + fn.Name)
	}
	evalContext.Variables["function"] = cty.ObjectVal(functionMap)

	filesMap := map[string]cty.Value{}
	for _, fn := range cfg.Files {
		filesMap[fn.Name] = cty.StringVal("files." + fn.Name)
	}
	evalContext.Variables["files"] = cty.ObjectVal(filesMap)

	// TODO: validate that static files are in the project directory

	d := gohcl.DecodeBody(file.Body, evalContext, cfg)
	if d.HasErrors() {
		err = d
		return
	}

	cfg.filesMap = make(map[string]Files)
	for _, file := range cfg.Files {
		cfg.filesMap[file.Name] = file
	}

	cfg.databaseMap = make(map[string]*Database)
	for _, db := range cfg.Databases {
		cfg.databaseMap[db.Name] = &db
	}
	return
}

// FileName is the name of the embly configuration file
var FileName = "embly.hcl"

// FindConfigFile recursively searches all parent directories for an embly.hcl file
func FindConfigFile(wd string) (f *os.File, location string, err error) {

	if !filepath.IsAbs(wd) {
		var currentWorkingDirectory string
		if currentWorkingDirectory, err = os.Getwd(); err != nil {
			err = errors.WithStack(err)
			return
		}
		wd = filepath.Join(currentWorkingDirectory, wd)
	}

	for {
		if f, err = os.Open(filepath.Join(wd, FileName)); err == nil {
			break
		}
		parent := filepath.Join(wd, "../")
		if wd == parent || wd == "/" {
			break
		}
		wd = parent
	}
	location = wd

	if f == nil {
		err = errors.Errorf("%s not found in this directory or any parent", FileName)
		return
	}
	return
}
