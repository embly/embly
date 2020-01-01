package config

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	f, _, err := FindConfigFile("")
	if err != nil {
		t.Error(err)
	}
	if _, err := ParseConfig(f); err != nil {
		t.Error(err)
	}
}
