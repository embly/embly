package config

import (
	"log"
	"os"
)

// TODO: ignore during tests?
// if flag.Lookup("test.v") == nil {
// 	fmt.Println("normal run")
// } else {
// 	fmt.Println("run under go test")
// }

var config = map[string]string{}

// Register config values
func Register(keys ...string) {
	for _, key := range keys {
		val := os.Getenv(key)
		if val == "" {
			log.Fatalf("Config value %s is not set", key)
		}
		config[key] = val
	}
}

// Get gets a config value
func Get(key string) (val string) {
	var ok bool
	if val, ok = config[key]; ok == false {
		log.Fatalf("Config value %s was never initialized", key)
	}
	return
}
