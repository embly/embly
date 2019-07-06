package cache

import (
	"github.com/gomodule/redigo/redis"
)

// Client is a cache client wrapped around redis
type Client interface {
	Set(key, val string) error
	Get(key string) (string, error)
	SetB(key string, val []byte) error
	GetB(key string) ([]byte, error)
}

var cachePrefix = "embly-go-"

// TODO: not thread safe: https://godoc.org/github.com/gomodule/redigo/redis#hdr-Concurrency
// Make it thread safe

// NewClient created a new cache client
func NewClient(conn redis.Conn) Client {
	c := client{conn}
	return &c
}

type client struct {
	conn redis.Conn
}

func set(c redis.Conn, key string, v interface{}) (err error) {
	if err = c.Err(); err != nil {
		return
	}
	c.Send("SET", key, v)
	if err = c.Flush(); err != nil {
		return
	}
	return
}

func get(c redis.Conn, key string) (v interface{}, err error) {
	if err = c.Err(); err != nil {
		return
	}
	c.Send("GET", key)
	if err = c.Flush(); err != nil {
		return
	}
	return c.Receive()
}

func (c *client) Set(key, val string) (err error) {
	return set(c.conn, key, val)
}
func (c *client) Get(key string) (string, error) {
	val, err := get(c.conn, key)
	if err != nil {
		return "", nil
	}
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}
func (c *client) SetB(key string, val []byte) error {
	return set(c.conn, key, val)
}

func (c *client) GetB(key string) ([]byte, error) {
	val, err := get(c.conn, key)
	if err != nil {
		return nil, nil
	}
	switch v := val.(type) {
	case []byte:
		return v, nil
	default:
		return nil, nil
	}
}
