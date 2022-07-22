package client

import (
	"testing"
	"time"

	"git.intra.123u.com/rometa/romq/msg"

	"git.intra.123u.com/rometa/romq/client"
	nlog "github.com/abc463774475/my_tool/n_log"
)

func TestClientSub(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	c := client.NewClient("localhost:8088")

	c.Subscribe("haorena", func(data []byte, _msg *msg.MsgPub) {
		nlog.Debug("data: %s", string(data))
	}, func(_msg *msg.MsgSubAck) {
		nlog.Debug("sub ack: %+v", _msg)
	})

	time.Sleep(100 * time.Second)
}

func TestClientPub(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	c := client.NewClient("localhost:8089")

	c.Publish("haorena", []byte("hello 111111111"))

	time.Sleep(100 * time.Second)
}
