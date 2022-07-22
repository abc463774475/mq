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
		nlog.Debug("data: %s", string(data))
	}, func(_msg *msg.MsgSubAck) {
		nlog.Debug("sub ack: %+v", _msg)
		if _msg.Code != 0 {
			nlog.Erro("sub ack: code error %+v", _msg.Code)
		}
	})

	time.Sleep(100 * time.Second)
}
