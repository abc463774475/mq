package gateway

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"git.intra.123u.com/rometa/romq/client"
	"git.intra.123u.com/rometa/romq/msg"
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
)

var c *client.Client

func playerLogin() {
	playerID := snowflake.GetID()
	str := fmt.Sprintf("gateway:%v", playerID)

	c.Subscribe(str, func(data []byte, pub *msg.MsgPub) {
		nlog.Debug("data: %s", string(data))
	}, func(_msg *msg.MsgSubAck) {
		nlog.Debug("sub ack: %+v", _msg)
		if _msg.Code != 0 {
			nlog.Erro("sub ack: code error %+v", _msg.Code)
		}
	})

	{
		data, _ := json.Marshal(msg.BaseMgrMsgPlayerLogin{
			PlayerID: playerID,
		})

		c.Req("baseMgr", msg.BaseMgrMsg{
			MsgID: msg.BaseMgrMsgID_PlayerLogin,
			Data:  data,
		}, func(data []byte) {
			nlog.Erro("recv PlayerLogin: %s", string(data))
			// 这里保存了 baseID 相关的信息, 消息路由的时候, 可以根据 baseID 进行路由
			// 同理 cell , 可以根据 cellID 进行路由
		})
	}
}

func TestGateway(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Easy))
	snowflake.Init(1)
	c = client.NewClient("localhost:8087")

	c.Subscribe("base:player:10001", func(data []byte, _msg *msg.MsgPub) {
	}, nil)

	playerLogin()

	time.Sleep(1000 * time.Second)
}

func sendMsg(msgID int, i interface{}) {
	// send to base
	if msgID > 10000 && msgID <= 20000 {
		c.Publish("base:player:10001", msg.BaseMsg{
			MsgID: msg.BaseMsgID(msgID),
		})
	}

	// send to cell
	if msgID > 20000 && msgID <= 30000 {
		c.Publish("cell:player:10001", msg.CellMsg{
			MsgID: msg.CellMsgID(msgID),
		})
	}
}
