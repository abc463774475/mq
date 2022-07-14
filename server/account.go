package server

import (
	"git.intra.123u.com/rometa/romq/msg"
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"strconv"
	"sync"
	"time"
)

const (
	globalAccountName = "global"
)

// Account based limits.
type limits struct {
	mpay   int32
	msubs  int32
	mconns int32
}

type Account struct {
	stats
	limits
	srv *server

	name     string
	uniqueID int64
	updated  time.Time
	created  time.Time

	client *client

	sl *sublist

	clients map[*client]struct{}

	rm map[string]int32

	rwmu sync.RWMutex
	sqmu sync.Mutex
}

func NewAccount(name string) *Account {
	a := &Account{}
	a.init()
	a.name = name
	a.uniqueID = snowflake.GetID()
	a.created = time.Now()
	a.updated = a.created
	a.limits = limits{
		mpay:   -1,
		msubs:  -1,
		mconns: -1,
	}
	a.sl = newSublist()
	return a
}

func (a *Account) init() {
	a.clients = make(map[*client]struct{})
	a.rm = make(map[string]int32)
}

func (a *Account) String() string {
	return a.name + strconv.FormatUint(uint64(a.uniqueID), 36)
}

func (a *Account) setCurClient(client2 *client) {
	a.client = client2
}

func (a *Account) addClient(c *client) {
	if _, ok := a.clients[c]; ok {
		nlog.Erro("Account.addClient: client already added", c)
		return
	}
	a.clients[c] = struct{}{}
	nlog.Debug("Account.addClient: added client %v", a.name)
}

func (a *Account) addRM(name string, value int32) {
	a.rwmu.Lock()
	if _, ok := a.rm[name]; ok {
		nlog.Erro("Account.addRM: name already added", name)
		a.rwmu.Unlock()
		return
	}
	a.rm[name] = value
	a.rwmu.Unlock()
}

func (a *Account) delRM(name string) {
	a.rwmu.Lock()
	if _, ok := a.rm[name]; !ok {
		nlog.Erro("Account.delRM: name not found", name)
		a.rwmu.Unlock()
		return
	}
	delete(a.rm, name)
	a.rwmu.Unlock()
}

func (a *Account) getMsgAccounts() msg.Accounts {
	a.rwmu.RLock()
	defer a.rwmu.RUnlock()
	rs := msg.Accounts{
		RM:   make(map[string]int32, len(a.rm)),
		Name: a.name,
	}
	for name, value := range a.rm {
		rs.RM[name] = value
	}
	return rs
}
