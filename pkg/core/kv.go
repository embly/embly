package core

import (
	"net"
	"strings"

	comms_proto "embly/pkg/core/proto"
	"embly/pkg/kv"

	"github.com/pkg/errors"
)

// KV is the send/recv context for the KV store
type KV struct {
	master   *Master
	id       uint64
	conn     net.Conn
	isGetter bool

	// the action path "connect" "request"
	path string
}

var store = kv.NewMemoryStore()

func (k *KV) processRequest(msg comms_proto.Message) (err error) {
	if k.isGetter {
		value, err := store.Get(msg.Data)
		if err != nil {
			return err
		}
		if err := WriteMessage(k.conn, comms_proto.Message{
			Data: value,
			From: msg.To,
			To:   msg.From,
		}); err != nil {
			return err
		}
	} else {
		key, value, err := kv.ExtractKeyAndValue(msg.Data)
		if err != nil {
			return err
		}
		if err := store.Set(key, value); err != nil {
			return err
		}
		if err := WriteMessage(k.conn, comms_proto.Message{
			From: msg.To,
			To:   msg.From,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (k *KV) sendMsg(msg comms_proto.Message) {
	if err := k.processRequest(msg); err != nil {
		_ = WriteMessage(k.conn, comms_proto.Message{
			Data:  []byte(err.Error()),
			From:  msg.To,
			To:    msg.From,
			Error: 28,
		})
	}
}

func (master *Master) spawnKV(msg comms_proto.Message, conn net.Conn) (err error) {
	parts := strings.Split(msg.Spawn, "/")
	if len(parts) < 3 {
		return errors.New("missing command on kv request, should be embly/kv/set or embly/kv/get")
	}

	path := parts[2]

	k := &KV{
		master: master,
		id:     msg.SpawnAddress,
		conn:   conn,
		path:   path,
	}

	if path == "get" {
		k.isGetter = true
	} else if path == "set" {
		k.isGetter = false
	} else {
		return errors.New("path is not a known command")
	}

	master.addFuncOrGateway(msg.SpawnAddress, k)
	return nil
}
