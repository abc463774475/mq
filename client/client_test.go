package client

import (
	"testing"
	"time"

	"git.intra.123u.com/rometa/romq/msg"
)

func TestClient_sub(t *testing.T) {
	sub := "haorena"

	c := newClient("localhost:8088", 0)
	time.AfterFunc(1*time.Second, func() {
		c.sendMsg(msg.MSG_SUB, msg.MsgSub{
			Sub: sub,
			SID: "1",
		})
	})

	// c.run()
}

func TestClient_pub(t *testing.T) {
	c := newClient("localhost:8089", 0)
	time.AfterFunc(2*time.Second, func() {
		c.sendMsg(msg.MSG_PUB, msg.MsgPub{
			Sub:  "haorena",
			Data: []byte("hello 111111111"),
		})
	})
	c.run()
}
