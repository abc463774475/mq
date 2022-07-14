package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"git.intra.123u.com/rometa/romq/msg"
	nlog "github.com/abc463774475/my_tool/n_log"
)

type Client struct {
	nc   net.Conn
	addr string

	clientType int
	msgRecv    chan *msg.Msg
	msgSend    chan *msg.Msg

	cquit chan struct{}
}

func newClient(addr string, ct int) *Client {
	c := &Client{
		nc:         nil,
		addr:       addr,
		clientType: ct,
		msgRecv:    make(chan *msg.Msg, 200),
		msgSend:    make(chan *msg.Msg, 200),
		cquit:      make(chan struct{}, 1),
	}

	return c
}

func (c *Client) readLoop() {
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
	for {
		select {
		case msg := <-c.msgSend:
			c.writeMsg(msg)
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
			return errors.New(fmt.Sprintf("writeMsg: write 0 bytes %v", n1))
		}

	}
	return nil
}

func (c *Client) closeConnection() {
	c.nc.Close()
}

func (c *Client) processMsg() {
	for {
		select {
		case msg := <-c.msgRecv:
			c.processMsgImpl(msg)
		}
	}
}

func (c *Client) run() {
	// Set the read deadline for the client.
	// Start reading.
	nlog.Info("client run")
	defer func() {
		nlog.Info("client run end")
		c.closeConnection()
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
}

func (c *Client) connect() {
	var err error
	c.nc, err = net.Dial("tcp", c.addr)
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
		c.SendMsg(msg.MSG_PONG, msg.MsgPong{})
	case msg.MSG_PONG:
	case msg.MSG_HANDSHAKE:
	case msg.MSG_SNAPSHOTSUBS:
	case msg.MSG_SUB:
	case msg.MSG_PUB:
		c.processMsgPub(_msg)
	}
}

func (c *Client) register() {
	nlog.Info("register")
	c.SendMsg(msg.MSG_HANDSHAKE, &msg.MsgHandshake{
		Type: int32(c.clientType),
		Name: "client test",
	})
}

func (c *Client) SendMsg(msgID msg.MSGID, i interface{}) {
	data, err := json.Marshal(i)
	if err != nil {
		nlog.Erro("SendMsg: json.Marshal: %v", err)
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
		nlog.Erro("SendMsg: msgSend is full")
	}
}

func (c *Client) processMsgPub(_msg *msg.Msg) {
	pub := &msg.MsgPub{}
	err := json.Unmarshal(_msg.Data, pub)
	if err != nil {
		nlog.Erro("processMsgPub: json.Unmarshal: %v", err)
		return
	}
	nlog.Debug("processMsgPub: %v  %v", pub.Topic, string(pub.Data))
}
