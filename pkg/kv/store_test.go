package kv

import (
	"bytes"
	"testing"

	"embly/pkg/randy"
)

func TestExtract(t *testing.T) {
	key := []byte("key")
	value := []byte("value")

	{
		out, err := WriteKeyAndValue(key, value)
		if err != nil {
			t.Fatal(err)
		}
		k, v, err := ExtractKeyAndValue(out)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(k, key) {
			t.Error(k, "doesn't equal", key)
		}
		if !bytes.Equal(v, value) {
			t.Error(v, "doesn't equal", value)
		}
	}

	{
		_, err := WriteKeyAndValue([]byte(randy.String(10001)), value)
		if err != ErrKeyTooLarge {
			t.Fatal("key should be too large")
		}
	}

	{
		_, err := WriteKeyAndValue(key, []byte(randy.String(100001)))
		if err != ErrValueTooLarge {
			t.Fatal("value should be too large")
		}
	}

	{
		_, _, err := ExtractKeyAndValue([]byte{})
		if err != errNoSpaceForByteSize {
			t.Fatal("should not find space for size bytes")
		}
	}

	{
		_, _, err := ExtractKeyAndValue([]byte{2, 0, 0})
		if err != errInvalidSize {
			t.Fatal("should not find space for size bytes")
		}
	}

	{

		k, _, err := ExtractKeyAndValue([]byte{1, 0, 0})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte{0}, k) {
			t.Fatal("should have found small key value")
		}
	}

}
