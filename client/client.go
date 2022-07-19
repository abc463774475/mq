package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"git.intra.123u.com/rometa/romq/msg"
	nlog "github.com/abc463774475/my_tool/n_log"
)

type (
	SUBFUN func(data []byte)
	Client struct {
		nc   net.Conn
		addr string

		clientType int
		msgRecv    chan *msg.Msg
		msgSend    chan *msg.Msg

		rttDuration time.Duration
		cquit       chan struct{}

		sfs     map[string]SUBFUN
		rwmuSFs sync.RWMutex

		sid uint64
	}
)

func newClient(addr string, ct int) *Client {
	c := &Client{
		nc:         nil,
		addr:       addr,
		clientType: ct,
		msgRecv:    make(chan *msg.Msg, 200),
		msgSend:    make(chan *msg.Msg, 200),
		cquit:      make(chan struct{}, 2),
		sfs:        make(map[string]SUBFUN),
	}

	go c.run()

	return c
}

func NewClient(addr string) *Client {
	return newClient(addr, 0)
}

func (c *Client) readLoop() {
	nlog.Info("readLoop")
	defer func() {
		nlog.Info("readLoop end")
	}()
	for {
		headeBytes := make([]byte, msg.HeadSize)
		n, err := io.ReadFull(c.nc, headeBytes)
		if err != nil || n == 0 {
			nlog.Erro("readLoop: read header: %v", err)
			c.closeConnection()
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
			c.closeConnection()
			return
		}

		c.msgRecv <- &msg.Msg{
			Head: *header,
			Data: data,
		}
	}
}

func (c *Client) writeLoop() {
	nlog.Info("writeLoop")
	defer func() {
		nlog.Info("writeLoop end")
	}()
	for {
		select {
		case msg := <-c.msgSend:
			_ = c.writeMsg(msg)
		case <-c.cquit:
			return
		}
	}
}

func (c *Client) writeMsg(msg *msg.Msg) error {
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
			return fmt.Errorf("writeMsg: write 0 bytes %v", n1)
		}

	}
	return nil
}

func (c *Client) closeConnection() {
	c.nc.Close()
	// just twice to make sure, since there have two goroutines
	c.cquit <- struct{}{}
	c.cquit <- struct{}{}
}

func (c *Client) del() {
	close(c.msgRecv)
	close(c.msgSend)
	close(c.cquit)
}

func (c *Client) processMsg() {
	nlog.Info("processMsg")
	tick := time.NewTicker(time.Second * 5)

	defer func() {
		nlog.Info("processMsg end")
		tick.Stop()
	}()

	for {
		select {
		case msg := <-c.msgRecv:
			c.processMsgImpl(msg)
		case <-tick.C:
			c.sendPing()
		case <-c.cquit:
			return
		}
	}
}

func (c *Client) run() {
	// Set the read deadline for the client.
	// Start reading.
	nlog.Info("client run")
	defer func() {
		nlog.Info("client run end")
	}()

	c.connect()

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
	c.del()
}

func (c *Client) connect() {
	var err error
	c.nc, err = net.DialTimeout("tcp", c.addr, time.Second*5)

	if err != nil {
		panic("err = " + err.Error())
	}

	nlog.Info("connect: %v", c.addr)
	c.register()
}

func (c *Client) processMsgImpl(_msg *msg.Msg) {
	nlog.Info("processMsgImpl: %v  %v", _msg.Head.ID, string(_msg.Data))
	switch _msg.ID {
	case msg.MSG_PING:
		c.processMsgPing(_msg)
	case msg.MSG_PONG:
		c.processMsgPong(_msg)
	case msg.MSG_HANDSHAKE:
	case msg.MSG_SUB:
	case msg.MSG_PUB:
		c.processMsgPub(_msg)
	}
}

func (c *Client) register() {
	nlog.Info("register")
	c.sendMsg(msg.MSG_HANDSHAKE, &msg.MsgHandshake{
		Type: int32(c.clientType),
		Name: "client test",
	})
}

func (c *Client) sendMsg(msgID msg.MSGID, i interface{}) {
	data, err := json.Marshal(i)
	if err != nil {
		nlog.Erro("sendMsg: json.Marshal: %v", err)
		return
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
		nlog.Erro("sendMsg: msgSend is full")
	}
}

func (c *Client) processMsgPub(_msg *msg.Msg) {
	pub := &msg.MsgPub{}
	err := json.Unmarshal(_msg.Data, pub)
	if err != nil {
		nlog.Erro("processMsgPub: json.Unmarshal: %v", err)
		return
	}
	nlog.Debug("processMsgPub: %v  %v", pub.Sub, string(pub.Data))

	c.rwmuSFs.RLock()
	sf, ok := c.sfs[pub.Sub]
	if !ok {
		c.rwmuSFs.RUnlock()
		nlog.Erro("processMsgPub: sub not found: %v", pub.Sub)
		return
	}
	c.rwmuSFs.RUnlock()

	sf(pub.Data)
}

func (c *Client) Subscribe(sub string, f SUBFUN) {
	c.rwmuSFs.Lock()
	c.sfs[sub] = f

	atomic.AddUint64(&c.sid, 1)
	sid := atomic.LoadUint64(&c.sid)
	strSID := strconv.FormatUint(sid, 20)
	c.sendMsg(msg.MSG_SUB, &msg.MsgSub{
		Sub: sub,
		SID: strSID,
	})

	c.rwmuSFs.Unlock()
}

func (c *Client) UnSubscribe(sub string) {
	c.rwmuSFs.Lock()
	if _, ok := c.sfs[sub]; !ok {
		nlog.Erro("UnSubscribe: %v not exist", sub)
		c.rwmuSFs.Unlock()
		return
	}

	c.sendMsg(msg.MSG_UNSUB, &msg.MsgUnSub{
		Subs: []string{sub},
	})
	delete(c.sfs, sub)
	c.rwmuSFs.Unlock()
}

func (c *Client) Publish(sub string, data []byte) {
	c.sendMsg(msg.MSG_PUB, &msg.MsgPub{
		Sub:  sub,
		Data: data,
	})
}
