package client

import (
	"testing"
	"time"

	"git.intra.123u.com/rometa/romq/msg"
)

func TestClient_sub(t *testing.T) {
	c := newClient("localhost:8088", 0)
	time.AfterFunc(2*time.Second, func() {
		c.SendMsg(msg.MSG_SUB, msg.MsgSub{
			Topic: "haorena",
			SID:   "1",
		})
	})
	c.run()
}

func TestClient_pub(t *testing.T) {
	c := newClient("localhost:8089", 0)
	time.AfterFunc(2*time.Second, func() {
		c.SendMsg(msg.MSG_PUB, msg.MsgPub{
			Topic: "haorena",
			Data:  []byte("hello 111111111"),
		})
	})
	c.run()
}
