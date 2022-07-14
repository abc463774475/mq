package msg

type MSGID int32

const (
	MSG_START MSGID = iota + 1
	MSG_PING
	MSG_PONG
	MSG_HANDSHAKE
	MSG_SNAPSHOTSUBS
	MSG_SUB
	MSG_PUB
	MSG_ROUTEPUB

	MSG_NEWROUTE
	MSG_REMOTEROUTEADDSUB
)

type MsgPing struct {
}

type MsgPong struct {
}

type MsgHandshake struct {
	Type int32  `json:"type"`
	Name string `json:"name"`
}

type MsgSnapshot struct {
	Data []byte `json:"data"`
}

type MsgSub struct {
	Topic string `json:"topic"`
	// SID sequence id
	SID string `json:"sid"`
}

type MsgPub struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}

type MsgSnapshotSubs struct {
	All []*Accounts `json:"all"`
}

type MsgNewRoute struct {
	Name string `json:"name"`
}

type Accounts struct {
	Name string           `json:"name"`
	RM   map[string]int32 `json:"rm"`
}

type MsgRemoteRouteAddSub struct {
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type MsgRoutePub struct {
	Topic string `json:"topic"`
	Data  []byte `json:"data"`
}
