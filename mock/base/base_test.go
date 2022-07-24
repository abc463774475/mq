package base

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

var (
	c *client.Client

	baseID  = 1
	subName = fmt.Sprintf("base:%d", baseID)
)

func processBaseMsg(_msg *msg.BaseMsg, pub *msg.MsgPub) {
	switch _msg.MsgID {
	case msg.BaseMsgID_PlayerLogin:
		{
			_login := &msg.BaseMsgPlayerLogin{}
			err := json.Unmarshal(_msg.Data, _login)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("player login: %+v", _login)

			// todo db mock can not support
			// how to send back to client
			c.Publish(fmt.Sprintf("gateway:%v", _login.PlayerID), "login ok")

			c.Subscribe(fmt.Sprintf("base:player:%v", _login.PlayerID), func(data []byte, pub *msg.MsgPub) {
				nlog.Info("base:player subscribe:  callback %+v", string(data))
			}, nil)

			// send to cellMgr
		}
	}
}

func TestBase(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Easy))
	snowflake.Init(2)
	c = client.NewClient("localhost:8087")

	data, _ := json.Marshal(msg.BaseMgrMsgBaseRegister{
		ID: int32(baseID),
	})

	c.Req("baseMgr", msg.BaseMgrMsg{
		MsgID: msg.BaseMgrMsgID_BaseRegister,
		Data:  data,
	}, func(data []byte) {
		nlog.Erro("base register:  callback %+v", string(data))
	})

	c.Subscribe(subName, func(data []byte, _msg *msg.MsgPub) {
		nlog.Erro("base subscribe:  callback %+v", string(data))
		bMsg := &msg.BaseMsg{}
		err := json.Unmarshal(data, bMsg)
		if err != nil {
			nlog.Erro("json unmarshal error: %s", err.Error())
			return
		}
		processBaseMsg(bMsg, _msg)
	}, func(_msg *msg.MsgSubAck) {
		nlog.Debug("sub ack: %+v", _msg)
		if _msg.Code != 0 {
			nlog.Erro("sub ack: code error %+v", _msg.Code)
		}
	})

	time.Sleep(10000 * time.Second)
}
