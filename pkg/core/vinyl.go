package core

import (
	"fmt"
	"net"
	"strings"
	"time"

	comms_proto "embly/pkg/core/proto"

	vinyl "github.com/embly/vinyl/vinyl-go"
	"github.com/embly/vinyl/vinyl-go/transport"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// Vinyl is the send/recv context for a vinyl call
type Vinyl struct {
	master   *Master
	id       uint64
	database string
	conn     net.Conn

	// the action path "connect" "request"
	path string
}

func (v *Vinyl) sendDBRequest(msg comms_proto.Message) (resp *transport.Response, err error) {
	t := time.Now()
	db := v.master.databases[v.database]
	request := transport.Request{}

	if err = proto.Unmarshal(msg.Data, &request); err != nil {
		return
	}
	resp, err = db.DB.SendRequest(request)
	v.master.ui.Output(
		fmt.Sprintf("Vinyl: %s (%s)",
			vinyl.RequestDescription(&request), time.Since(t)))
	return
}

func (v *Vinyl) sendMsg(msg comms_proto.Message) {
	go func() {
		resp, err := v.sendDBRequest(msg)
		if err != nil {
			resp = &transport.Response{
				Error: err.Error(),
			}
		}
		data, _ := proto.Marshal(resp)
		if err := WriteMessage(v.conn, comms_proto.Message{
			Data: data,
			From: v.id,
			To:   msg.From,
		}); err != nil {
			v.master.ui.Error("Error sending message back to function " + err.Error())
		}
	}()
}

func (master *Master) spawnVinyl(msg comms_proto.Message, conn net.Conn) (err error) {
	parts := strings.Split(msg.Spawn, "/")
	if len(parts) <= 2 {
		return errors.New("missing database name")
	}
	database := parts[2]
	if len(parts) <= 3 {
		return errors.New("missing database path")
	}
	path := parts[3]
	if path == "connect" {
		db := master.databases[database]
		if err := WriteMessage(
			conn,
			comms_proto.Message{
				Data: []byte(db.DB.Token),
				From: msg.SpawnAddress,
				To:   msg.From,
			},
		); err != nil {
			master.ui.Error(err.Error())
			return err
		}
		return nil
	}
	if path != "request" {
		return errors.New("vinyl path doesn't match 'request'")
	}

	v := &Vinyl{
		master:   master,
		id:       msg.SpawnAddress,
		database: database,
		conn:     conn,
		path:     path,
	}

	master.addFuncOrGateway(msg.SpawnAddress, v)
	return nil
}
