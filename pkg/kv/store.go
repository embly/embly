package kv

import (
	"encoding/binary"
	"errors"
)

// Store is the interface for getting and setting
type Store interface {
	Get(key []byte) (value []byte, err error)
	Set(key []byte, value []byte) (err error)
}

var (
	// ErrNoExist the requested value doesn't exist
	ErrNoExist = errors.New("error doesn't exist")

	errNoSpaceForByteSize = errors.New("invalid input length for key/value string")
	errInvalidSize        = errors.New("key length is longer than input string")

	// ErrKeyTooLarge key is too large
	ErrKeyTooLarge = errors.New("key values can't be greater than 10,000")

	// ErrValueTooLarge value is too large
	ErrValueTooLarge = errors.New("value can't be greater than 100,000")
)

// ExtractKeyAndValue will take input bytes that are structured like so:
// |uint16 len of key|key bytes|value bytes|
// will error if bytes can't be parsed
func ExtractKeyAndValue(input []byte) (key []byte, value []byte, err error) {
	if len(input) < 2 {
		err = errNoSpaceForByteSize
		return
	}
	ln := binary.LittleEndian.Uint16(input[:2])
	if int(ln) > len(input[2:]) {
		err = errInvalidSize
		return
	}
	return input[2 : 2+ln], input[ln+2:], nil
}

// WriteKeyAndValue will write key and value bytes to a contiguous byte array
// that is prefixed with uint16 little endian bytes of the key size so that they
// can be parsed out elsewhere
func WriteKeyAndValue(key []byte, value []byte) (out []byte, err error) {
	if len(key) > 10000 {
		err = ErrKeyTooLarge
		return
	}
	if len(value) > 100000 {
		err = ErrValueTooLarge
		return
	}

	out = []byte{0, 0}
	keyLen := uint16(len(key))
	binary.LittleEndian.PutUint16(out, keyLen)
	out = append(out, key...)
	out = append(out, value...)
	return
}
