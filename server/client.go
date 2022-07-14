package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.intra.123u.com/rometa/romq/msg"
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"io"
	"net"
	"sync"
	"time"
)

type ClientType int

// type of client connection
const (
	CLIENT ClientType = iota
	ROUTER
	SYSTEM
	ACCOUNT
)

type msgDeny struct {
	reason string
	deny   *sublist
	dcache map[string]bool
}

const (
	maxBuffersSize = 1024 * 1024 * 64
)

type clientFlag int

const (
	handshake clientFlag = 1 << iota
	readOnly
	expectConnect
	closeConnection
)

func (cf *clientFlag) set(c clientFlag) {
	*cf |= c
}

func (cf *clientFlag) clear(c clientFlag) {
	*cf &= ^c
}

func (cf clientFlag) isSet(c clientFlag) bool {
	return cf&c != 0
}

func (cf *clientFlag) setIfNotSet(c clientFlag) bool {
	if *cf&c == 0 {
		*cf |= c
		return true
	}
	return false
}

type closeState int

const (
	clientClosed closeState = iota + 1
	writeError
	readError
)

type readCache struct {
	id      uint64
	results map[string]*sublistResult

	msgs  int64
	bytes int64
	subs  int32

	rsz int32
	srs int32
}

type routeTarget struct {
	sub *subscription
	qs  []byte
	_qs [32]byte
}

type outbound struct {
	p   []byte        // Primary write buffer
	s   []byte        // Secondary for use post flush
	nb  net.Buffers   // net.Buffers for writev IO
	sz  int32         // limit size per []byte, uses variable BufSize constants, start, min, max.
	sws int32         // Number of short writes, used for dynamic resizing.
	pb  int64         // Total pending/queued bytes.
	pm  int32         // Total pending/queued messages.
	fsp int32         // Flush signals that are pending per producer from readLoop's pcd.
	sg  *sync.Cond    // To signal writeLoop that there is data to flush.
	wdl time.Duration // Snapshot of write deadline.
	mp  int64         // Snapshot of max pending for client.
	lft time.Duration // Last flush time for Write.
	stc chan struct{} // Stall chan we create to slow down producers on overrun, e.g. fan-in.
}

type client struct {
	id             int64
	kind           ClientType
	isServerAccept bool

	name string

	stats
	srv *server
	acc *Account

	mpay  int32
	msubs int32

	nc net.Conn

	in  readCache
	out outbound

	addr string

	subs   map[string]*subscription
	mperms *msgDeny

	rtt      time.Duration
	rttStart time.Time

	msgRecv chan *msg.Msg
	msgSend chan *msg.Msg

	cquit chan struct{}

	mu sync.Mutex
}

func newAcceptClient(id int64, conn net.Conn, s *server) *client {
	c := &client{}
	c.id = id
	c.srv = s
	c.nc = conn
	c.isServerAccept = true

	return c
}

func newConnectClient(addr string) *client {
	c := &client{}
	c.id = snowflake.GetID()
	c.addr = addr
	c.isServerAccept = false
	return c
}

func (c *client) init() {
	c.subs = make(map[string]*subscription)
	c.mperms = &msgDeny{}
	c.msgRecv = make(chan *msg.Msg, 200)
	c.msgSend = make(chan *msg.Msg, 200)
	c.cquit = make(chan struct{}, 1)
}

func (c *client) readLoop() {
	for {
		headeBytes := make([]byte, msg.HeadSize)
		n, err := io.ReadFull(c.nc, headeBytes)
		if err != nil || n == 0 {
			nlog.Erro("readLoop: read header: %v", err)
			c.closeConnection(readError)
			return
		}

		header := &msg.Head{}
		header.Load(headeBytes)

		if header.Length == 0 {
			c.msgRecv <- &msg.Msg{
				Head: *header,
				Data: []byte{},
			}
			continue
		}

		data := make([]byte, header.Length)
		n, err = io.ReadFull(c.nc, data)
		if err != nil || n == 0 {
			nlog.Erro("readLoop: read data: %v", err)
			c.closeConnection(readError)
			return
		}

		c.msgRecv <- &msg.Msg{
			Head: *header,
			Data: data,
		}
	}
}

func (c *client) writeLoop() {
	for {
		select {
		case msg := <-c.msgSend:
			c.writeMsg(msg)
		case <-c.cquit:
			return
		}
	}
}

func (c *client) writeMsg(msg *msg.Msg) error {
	data := msg.Save()
	n, err := c.nc.Write(data)
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("writeMsg: write 0 bytes")
	}

	if n != len(data) {
		time.Sleep(1 * time.Second)
		n1, err := c.nc.Write(data[n:])
		if err != nil {
			return err
		}

		if n1 != len(data[n:]) {
			return errors.New(fmt.Sprintf("writeMsg: write 0 bytes %v", n1))
		}

	}
	return nil
}

func (c *client) closeConnection(state closeState) {
	nlog.Info("closeConnection: %v", state)
	c.nc.Close()
}

func (c *client) processMsg() {
	for {
		select {
		case msg := <-c.msgRecv:
			c.processMsgImpl(msg)
		}
	}
}

func (c *client) run() {
	// Set the read deadline for the client.
	// Start reading.
	nlog.Info("client run")
	defer func() {
		nlog.Info("client run end")
		c.closeConnection(clientClosed)
	}()

	//if c.isServerAccept {
	//	c.connect()
	//}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		c.readLoop()
		wg.Done()
	}()
	go func() {
		c.writeLoop()
		wg.Done()
	}()

	c.processMsg()

	wg.Wait()
}

func (c *client) SendMsg(msgID msg.MSGID, i interface{}) {
	var data []byte
	var err error
	if d, ok := i.([]byte); ok {
		data = d
	} else if d, ok := i.(string); ok {
		data = []byte(d)
	} else {
		data, err = json.Marshal(i)
		if err != nil {
			nlog.Erro("SendMsg: json.Marshal: %v", err)
			return
		}
	}

	msg := &msg.Msg{
		Head: msg.Head{
			ID:      msgID,
			Length:  0,
			Crc32:   0,
			Encrypt: 0,
		},
		Data: data,
	}

	if len(c.msgSend) < cap(c.msgSend) {
		c.msgSend <- msg
	} else {
		nlog.Erro("SendMsg: msgSend is full")
	}
}

func (c *client) Ping() {
	c.rttStart = time.Now()
	c.SendMsg(msg.MSG_PING, msg.MsgPing{})
}

func (c *client) processMsgHandshake(_msg *msg.Msg) {
	handshake := &msg.MsgHandshake{}
	err := json.Unmarshal(_msg.Data, handshake)
	if err != nil {
		nlog.Erro("processMsgHandshake: json.Unmarshal: %v", err)
		c.closeConnection(readError)
		return
	}

	c.name = handshake.Name
	c.kind = ClientType(handshake.Type)
	if handshake.Type == int32(CLIENT) {

	} else if handshake.Type == int32(ROUTER) {
		c.srv.addRoute(c, handshake.Name)
	} else if handshake.Type == int32(SYSTEM) {

	} else if handshake.Type == int32(ACCOUNT) {

	} else {
		nlog.Erro("processMsgHandshake: unknown type: %v", handshake.Type)
		return
	}
}

func (c *client) processMsgImpl(_msg *msg.Msg) {
	nlog.Info("processMsgImpl: %v  data %v", _msg.Head.ID, string(_msg.Data))
	switch _msg.ID {
	case msg.MSG_PING:
		c.SendMsg(msg.MSG_PONG, msg.MsgPong{})
	case msg.MSG_PONG:
		c.rtt = time.Now().Sub(c.rttStart)
		nlog.Info("processMsgImpl: %v  rtt %v", _msg.Head.ID, c.rtt)
	case msg.MSG_HANDSHAKE:
		c.processMsgHandshake(_msg)
	case msg.MSG_SNAPSHOTSUBS:
		c.processMsgSnapshotSubs(_msg)
	case msg.MSG_SUB:
		c.processMsgSub(_msg)
	case msg.MSG_PUB:
		c.processMsgPub(_msg)
	case msg.MSG_REMOTEROUTEADDSUB:
		c.processMsgRemoteRouteAddSub(_msg)
	case msg.MSG_ROUTEPUB:
		c.processMsgRoutePub(_msg)
	case msg.MSG_REGISTERROUTER:
		c.processMsgRegisterRouter(_msg)
	case msg.MSG_CURALLROUTES:
		c.processMsgCurAllRoutes(_msg)
	}
}

func (c *client) processMsgSnapshotSubs(_msg *msg.Msg) {
	snapshotSubs := &msg.MsgSnapshotSubs{}
	err := json.Unmarshal(_msg.Data, snapshotSubs)
	if err != nil {
		nlog.Erro("processMsgSnapshotSubs: json.Unmarshal: %v", err)
		c.closeConnection(readError)
		return
	}

	c.srv.snapshotSubs(c, snapshotSubs)
}

func (c *client) processMsgRoutePub(_msg *msg.Msg) {
	routePub := &msg.MsgRoutePub{}
	err := json.Unmarshal(_msg.Data, routePub)
	if err != nil {
		nlog.Erro("processMsgRoutePub: json.Unmarshal: %v", err)
		c.closeConnection(readError)
		return
	}

	c.msgPub(&msg.MsgPub{
		Topic: routePub.Topic,
		Data:  routePub.Data,
	})
}

func (c *client) processMsgSub(_msg *msg.Msg) {
	msub := &msg.MsgSub{}
	err := json.Unmarshal(_msg.Data, msub)
	if err != nil {
		nlog.Erro("processMsgSub: json.Unmarshal: %v", err)
		c.closeConnection(readError)
		return
	}

	sub := &subscription{
		client:  c,
		subject: msub.Topic,
		queue:   nil,
		qw:      0,
		closed:  0,
	}
	acc := c.acc
	c.mu.Lock()
	c.in.subs++
	sid := msub.SID

	ts := c.subs[sid]
	if ts == nil {
		c.subs[sid] = sub
		acc.addRM(sub.subject, 1)
		err := acc.sl.Insert(sub)
		if err != nil {
			delete(c.subs, sid)
		}
	}

	srv := c.srv
	c.mu.Unlock()

	kind := c.kind
	if kind == CLIENT {
		srv.updateRouteSubscriptionMap(acc, sub)
	}
}

func (c *client) processMsgPub(_msg *msg.Msg) {
	pub := &msg.MsgPub{}
	err := json.Unmarshal(_msg.Data, pub)
	if err != nil {
		nlog.Erro("processMsgPub: json.Unmarshal: %v", err)
		c.closeConnection(readError)
		return
	}

	c.msgPub(pub)
}

func (c *client) msgPub(pub *msg.MsgPub) {
	r := c.acc.sl.match(pub.Topic)
	if r == nil {
		nlog.Erro("processMsgPub: match not exist: %v", pub.Topic)
		return
	}

	for _, sub := range r.subs {
		if sub.client.kind == CLIENT {
			sub.client.SendMsg(msg.MSG_PUB, pub)
		} else if sub.client.kind == ROUTER {
			sub.client.SendMsg(msg.MSG_REMOTEROUTEADDSUB, &msg.MsgRoutePub{
				Topic: pub.Topic,
				Data:  pub.Data,
			})
		}
	}
}

func (c *client) registerWithAccount(acc *Account) error {
	nlog.Erro("client id %v registerWithAccount: %v", c.id, acc.name)
	c.acc = acc
	acc.addClient(c)
	return nil
}

func (c *client) sendRemoteNewSub(sub *subscription) {
	nlog.Erro("sendRemoteNewSub: %v", sub.subject)
	c.SendMsg(msg.MSG_REMOTEROUTEADDSUB, &msg.MsgRemoteRouteAddSub{
		Name:  c.name,
		Topic: sub.subject,
	})
}

func (c *client) processMsgRemoteRouteAddSub(_msg *msg.Msg) {
	remoteNewSub := &msg.MsgRemoteRouteAddSub{}
	err := json.Unmarshal(_msg.Data, remoteNewSub)
	if err != nil {
		nlog.Erro("processRemoteNewSub: json.Unmarshal: %v", err)
		// c.closeConnection(readError)
		return
	}

	acc := c.acc
	if acc == nil {
		nlog.Erro("processRemoteNewSub: acc is nil")
		return
	}

	srv := c.srv
	srv.lock.Lock()
	acc.addRM(remoteNewSub.Topic, 1)
	err = acc.sl.Insert(&subscription{
		client:  c,
		subject: remoteNewSub.Topic,
		queue:   nil,
		qw:      0,
		closed:  0,
	})
	srv.lock.Unlock()
}

func (c *client) processMsgRegisterRouter(_msg *msg.Msg) {
	registerRouter := &msg.MsgRegisterRouter{}
	err := json.Unmarshal(_msg.Data, registerRouter)
	if err != nil {
		nlog.Erro("processMsgRegisterRouter: json.Unmarshal: %v", err)
		return
	}
	c.srv.addRouterInfo(c, registerRouter)

	all := c.srv.getAllRouteInfos()
	retMsg := &msg.MsgCurAllRoutes{}
	for _, route := range all {
		if route.ID == registerRouter.Name {
			continue
		}
		retMsg.All = append(retMsg.All, &msg.RouterInfo{
			Name:        route.ID,
			ClientAddr:  route.ListenAddr,
			ClusterAddr: route.ClusterAddr,
		})
	}
	// 同时要回信息
	c.SendMsg(msg.MSG_CURALLROUTES, retMsg)
}

func (c *client) processMsgCurAllRoutes(_msg *msg.Msg) {
	curAllRoutes := &msg.MsgCurAllRoutes{}
	err := json.Unmarshal(_msg.Data, curAllRoutes)
	if err != nil {
		nlog.Erro("processMsgCurAllRoutes: json.Unmarshal: %v", err)
		return
	}
	c.srv.addRouterInfos(curAllRoutes.All)
}
