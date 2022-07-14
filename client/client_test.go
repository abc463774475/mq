package client

import (
	"git.intra.123u.com/rometa/romq/msg"
	"testing"
	"time"
)

func TestClient_common(t *testing.T) {
	c := newClient("localhost:8087", 0)
	time.AfterFunc(2*time.Second, func() {
		c.SendMsg(msg.MSG_SUB, msg.MsgSub{
			Topic: "haorena",
			SID:   "1",
		})
	})
	c.run()
}

func TestClient_router(t *testing.T) {
	c := newClient("localhost:8087", 1)
	//time.AfterFunc(2*time.Second, func() {
	//	c.SendMsg(msg.MSG_SUB, msg.MsgSub{
	//		Topic: "haorena",
	//		SID:   "1",
	//	})
	//})
	c.run()
}

func TestClient_sub(t *testing.T) {
	c := newClient("localhost:8087", 0)
	time.AfterFunc(2*time.Second, func() {
		c.SendMsg(msg.MSG_SUB, msg.MsgSub{
			Topic: "haorena",
			SID:   "1",
		})
	})
	c.run()
}

func TestClient_pub(t *testing.T) {
	c := newClient("localhost:8087", 0)
	time.AfterFunc(2*time.Second, func() {
		c.SendMsg(msg.MSG_PUB, msg.MsgPub{
			Topic: "haorena",
			Data:  []byte("hello 111111111"),
		})
	})
	c.run()
}
