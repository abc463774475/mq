package msg

import "time"

type MSGID int32

//go:generate go run github.com/dmarkham/enumer -type=MSGID
const (
	MSG_START MSGID = iota + 1
	MSG_PING
	MSG_PONG
	MSG_HANDSHAKE

	MSG_REGISTERROUTER

	MSG_SNAPSHOTSUBS

	MSG_SUB
	MSG_UNSUB

	MSG_PUB
	MSG_ROUTEPUB

	MSG_NEWROUTE
	MSG_REMOTEROUTEADDSUB
	MSG_REMOTEROUTEADDUNSUB

	MSG_CURALLROUTES
)

type MsgPing struct {
	Time time.Time `json:"time"`
}

type MsgPong struct{}

type MsgHandshake struct {
	Type int32  `json:"type"`
	Name string `json:"name"`
}

type MsgSnapshot struct {
	Data []byte `json:"data"`
}

type MsgSub struct {
	Sub string `json:"sub"`
	// SID sequence id
	SID string `json:"sid"`
}

type MsgUnSub struct {
	Subs []string `json:"subs"`
}

type MsgPub struct {
	Sub  string `json:"sub"`
	Data []byte `json:"data"`
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
	Name string   `json:"name"`
	Subs []string `json:"subs"`
}

type MsgRoutePub struct {
	Sub  string `json:"sub"`
	Data []byte `json:"data"`
}

type RouterInfo struct {
	Name        string `json:"name"`
	ClientAddr  string `json:"clientAddr"`
	ClusterAddr string `json:"clusterAddr"`
}

type MsgRegisterRouter struct {
	RouterInfo
}

type MsgCurAllRoutes struct {
	RemoteName string        `json:"remoteName"`
	All        []*RouterInfo `json:"all"`
}

type MsgRemoteRouteAddUnsub struct {
	Subs []string `json:"subs"`
}
