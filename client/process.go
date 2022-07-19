package client

import (
	"git.intra.123u.com/rometa/romq/msg"
)

func (c *Client) sendPing() {
	//c.sendMsg(msg.MSG_PING, msg.MsgPing{
	//	Time: time.Now(),
	//})
}

func (c *Client) processMsgPing(_msg *msg.Msg) {
	c.sendMsg(msg.MSG_PONG, msg.MsgPong{})
}

func (c *Client) processMsgPong(_msg *msg.Msg) {
}
