package msg

type BaseMgrMsgID int32

//go:generate go run github.com/dmarkham/enumer -type=BaseMgrMsgID
const (
	BaseMgrMsgID_Start BaseMgrMsgID = iota + 1
	BaseMgrMsgID_BaseRegister
	BaseMgrMsgID_PlayerLogin
)

type BaseMgrMsg struct {
	MsgID BaseMgrMsgID `json:"msgID"`
	Data  []byte       `json:"data"`
}

type BaseMgrMsgBaseRegister struct {
	ID int32 `json:"id"`
}

type BaseMgrMsgPlayerLogin struct {
	PlayerID int64 `json:"playerID"`
}

type BaseMgrMsgPlayerLoginAck struct {
	BaseID int32 `json:"baseID"`
}

type BaseMsgID int32

//go:generate go run github.com/dmarkham/enumer -type=BaseMsgID
const (
	BaseMsgID_Start BaseMsgID = iota + 1
	BaseMsgID_PlayerLogin

	BaseMsgID_End
)

type BaseMsg struct {
	MsgID BaseMsgID `json:"msgID"`
	Data  []byte    `json:"data"`
}

type BaseMsgPlayerLogin struct {
	PlayerID int64 `json:"playerID"`
}

type CellMgrMsgID int32

//go:generate go run github.com/dmarkham/enumer -type=CellMgrMsgID
const (
	CellMgrMsgID_Start CellMgrMsgID = iota + 1
	CellMgrMsgID_CellRegister
	CellMgrMsgID_PlayerLogin
	CellMgrMsgID_PlayerChangeCell

	CellMgrMsgID_End
)

type CellMgrMsg struct {
	MsgID CellMgrMsgID `json:"msgID"`
	Data  []byte       `json:"data"`
}

type CellMgrMsgCellRegister struct {
	ID int32 `json:"id"`
}

type CellMgrMsgPlayerLogin struct {
	PlayerID int64 `json:"playerID"`
}

type CellMgrMsgPlayerChangeCell struct {
	PlayerID int64 `json:"playerID"`
	CellID   int32 `json:"cellID"`
}

type CellMsgID int32

//go:generate go run github.com/dmarkham/enumer -type=CellMsgID
const (
	CellMsgID_Start CellMsgID = iota + 1
	CellMsgID_PlayerLogin
	CellMsgID_PlayerLogout

	CellMsgID_End
)

type CellMsg struct {
	MsgID CellMsgID `json:"msgID"`
	Data  []byte    `json:"data"`
}

type CellMsgPlayerLogin struct {
	PlayerID int64 `json:"playerID"`
}

type CellMsgPlayerLogout struct {
	PlayerID int64 `json:"playerID"`
}
