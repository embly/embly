package kv

import (
	"bytes"
	"testing"
)

func TestMemorySetAndGet(t *testing.T) {
	ms := NewMemoryStore()

	{
		key := []byte("key")
		value := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		ms.Set(key, value)

		v, err := ms.Get(key)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(v, value) {
			t.Fatal("bytes aren't equal")
		}
	}

	{
		v, err := ms.Get([]byte("noexist"))
		if v != nil {
			t.Fatal("value should be nil")
		}
		if err != ErrNoExist {
			t.Fatal("error should be ErrNoExist")
		}
	}
}
