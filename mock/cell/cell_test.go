package cell

import (
	"encoding/json"
	"fmt"
	"git.intra.123u.com/rometa/romq/client"
	"git.intra.123u.com/rometa/romq/msg"
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"testing"
)

var (
	c       *client.Client
	cellID  = 1
	sbuName = fmt.Sprintf("cell:%d", cellID)
)

func processCellMsg(_msg *msg.CellMsg, pub *msg.MsgPub) {
	switch _msg.MsgID {
	case msg.CellMsgID_PlayerLogin:
		{
			_login := &msg.CellMsgPlayerLogin{}
			err := json.Unmarshal(_msg.Data, _login)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("player login: %+v", _login)

			if pub.UniqueID != 0 {
				c.Publish(fmt.Sprintf("%v", pub.UniqueID), "login ok")
			}
		}
	}
}

func TestCell(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Easy))
	snowflake.Init(5)
	c = client.NewClient("localhost:8087")

	data, _ := json.Marshal(msg.CellMgrMsgCellRegister{
		ID: int32(cellID),
	})

	c.Req("cellMgr", msg.CellMgrMsg{
		MsgID: msg.CellMgrMsgID_CellRegister,
		Data:  data,
	}, func(data []byte) {
		nlog.Debug("cell register callback: %+v", string(data))
	})

	c.Subscribe(sbuName, func(data []byte, _msg *msg.MsgPub) {
		cm := &msg.CellMsg{}
		err := json.Unmarshal(data, cm)
		if err != nil {
			nlog.Erro("json unmarshal error: %s", err.Error())
			return
		}

		processCellMsg(cm, _msg)
	}, nil)
}

func sendToBase(playerID int64, id msg.BaseMsgID, i interface{}) {
	var data []byte
	switch i.(type) {
	case string:
		data = []byte(i.(string))
	case []byte:
		data = i.([]byte)
	case *string:
		data = []byte(*i.(*string))
	case *[]byte:
		data = *i.(*[]byte)
	default:
		data, _ = json.Marshal(i)
	}

	c.Publish(fmt.Sprintf("base:%v", playerID), msg.BaseMsg{
		MsgID: id,
		Data:  data,
	})
}
