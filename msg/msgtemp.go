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
