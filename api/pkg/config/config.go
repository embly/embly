package config

import (
	"flag"
	"log"
	"os"
)

var testing bool
func init() {
	if flag.Lookup("test.v") != nil {
		testing = true
	}
}

var config = map[string]string{}

// Register config values
func Register(keys ...string) {
	for _, key := range keys {
		val := os.Getenv(key)
		if val == "" && !testing {
			log.Fatalf("Config value %s is not set", key)
		}
		config[key] = val
	}
}

// Get gets a config value
func Get(key string) (val string) {
	var ok bool
	if val, ok = config[key]; ok == false && !testing {
		log.Fatalf("Config value %s was never initialized", key)
	}
	return
}
