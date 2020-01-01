package command

import (
	"testing"
)

func TestAllCommandDescs(t *testing.T) {
	c := createCLI()
	for _, v := range c.Commands {
		c, err := v()
		if err != nil {
			t.Error(err)
		}
		c.Help()
		c.Synopsis()
	}
}
