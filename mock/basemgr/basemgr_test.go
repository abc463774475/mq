package basemgr

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
	c       *client.Client
	sbuName = `baseMgr`

	allBases = map[int32]int{}
)

func processBaseMgrMsg(_msg *msg.BaseMgrMsg, pub *msg.MsgPub) {
	switch _msg.MsgID {
	case msg.BaseMgrMsgID_BaseRegister:
		{
			_register := &msg.BaseMgrMsgBaseRegister{}
			err := json.Unmarshal(_msg.Data, _register)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("base register: %+v", _register)
			allBases[_register.ID] = 1
			if pub.UniqueID != 0 {
				c.Publish(fmt.Sprintf("%v", pub.UniqueID), "register ok")
			}
		}
	case msg.BaseMgrMsgID_PlayerLogin:
		{
			_login := &msg.BaseMgrMsgPlayerLogin{}
			err := json.Unmarshal(_msg.Data, _login)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("player login: %+v", _login)

			var _baseID int32
			for k := range allBases {
				_baseID = k
			}

			if _baseID == 0 {
				nlog.Erro("no base")
				return
			}
			// how to send back to client
			if pub.UniqueID != 0 {
				c.Publish(fmt.Sprintf("%v", pub.UniqueID), msg.BaseMgrMsgPlayerLoginAck{
					BaseID: _baseID,
				})
			}
		}
	}
}

func TestBaseMgr(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Easy))
	snowflake.Init(2)
	c = client.NewClient("localhost:8087")

	c.Subscribe(sbuName, func(data []byte, pub *msg.MsgPub) {
		_msg := &msg.BaseMgrMsg{}
		err := json.Unmarshal(data, _msg)
		if err != nil {
			nlog.Erro("json unmarshal error: %s", err.Error())
			return
		}

		processBaseMgrMsg(_msg, pub)
	}, func(_msg *msg.MsgSubAck) {
		nlog.Debug("sub ack: %+v", _msg)
		if _msg.Code != 0 {
			nlog.Erro("sub ack: code error %+v", _msg.Code)
		}
	})

	time.Sleep(100 * time.Second)
}
