package randy

import (
	"crypto/rand"
	"fmt"
)

// String is a random enough string
func String(length ...int) string {
	l := 16
	if len(length) > 0 {
		l = length[0]
	}
	b := make([]byte, l)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
