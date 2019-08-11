package localbuild

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestBindingsFile(t *testing.T) {
	loc, err := writeBindingsFile()
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadFile(loc)
	if err != nil {
		t.Error(err)
	}
	bo, err := json.Marshal(bindings)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(bo, b) {
		t.Error(string(bo), string(b))
	}

	if err = os.Remove(loc); err != nil {
		t.Error(err)
	}
}
