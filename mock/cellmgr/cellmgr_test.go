package cellmgr

import (
	"encoding/json"
	"fmt"
	"git.intra.123u.com/rometa/romq/client"
	"git.intra.123u.com/rometa/romq/msg"
	nlog "github.com/abc463774475/my_tool/n_log"
	"testing"
)

var (
	c       *client.Client
	subName = "cellMgr"

	allCells = map[int32]int{}

	allPlayers = map[int64]struct {
		CellID   int32
		StatTime int64
	}{}
)

func processCellMgrMsg(_msg *msg.CellMgrMsg, pub *msg.MsgPub) {
	switch _msg.MsgID {
	case msg.CellMgrMsgID_CellRegister:
		{
			_register := &msg.CellMgrMsgCellRegister{}
			err := json.Unmarshal(_msg.Data, _register)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("cell register: %+v", _register)
			allCells[_register.ID] = 1
			if pub.UniqueID != 0 {
				c.Publish(fmt.Sprintf("%v", pub.UniqueID), "register ok")
			}
		}
	case msg.CellMgrMsgID_PlayerLogin:
		{
			_login := &msg.CellMgrMsgPlayerLogin{}
			err := json.Unmarshal(_msg.Data, _login)
			if err != nil {
				nlog.Erro("json unmarshal error: %s", err.Error())
				return
			}

			nlog.Debug("player login: %+v", _login)

			var _cellID int32
			for k := range allCells {
				_cellID = k
			}

			if _cellID == 0 {
				nlog.Erro("no cell")
				return
			}
			// how to send back to client
			if pub.UniqueID != 0 {
				c.Publish(fmt.Sprintf("%v", pub.UniqueID), "login ok")
			}
		}
	case msg.CellMgrMsgID_PlayerChangeCell:
		pcc := &msg.CellMgrMsgPlayerChangeCell{}
		err := json.Unmarshal(_msg.Data, pcc)
		if err != nil {
			nlog.Erro("json unmarshal error: %s", err.Error())
			return
		}

		nlog.Debug("player change cell: %+v", pcc)

		// todo old cell player leave, new cell player enter
		var data []byte
		{
			data, err = json.Marshal(msg.CellMsgPlayerLogout{
				PlayerID: pcc.PlayerID})
		}
		c.Req(fmt.Sprintf("cell:player:%v", pcc.PlayerID), msg.CellMsg{
			MsgID: msg.CellMsgID_PlayerLogout,
			Data:  data,
		}, func(data []byte) {
			var d1 []byte
			d1, _ = json.Marshal(msg.CellMsgPlayerLogin{
				PlayerID: pcc.PlayerID,
			})

			c.Publish(fmt.Sprintf("cell:player:%v", pcc.PlayerID), msg.CellMsg{
				MsgID: msg.CellMsgID_PlayerLogin,
				Data:  d1,
			})
		})
	}
}

func TestCellMgr(t *testing.T) {
	c = client.NewClient("localhost:8087")
	c.Subscribe(subName, func(data []byte, _msg *msg.MsgPub) {
		nlog.Debug("cellmgr: %+v", string(data))
		cm := &msg.CellMgrMsg{}
		err := json.Unmarshal(data, cm)
		if err != nil {
			nlog.Erro("json unmarshal error: %s", err.Error())
			return
		}
		processCellMgrMsg(cm, _msg)
	}, nil)
}
