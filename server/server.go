package server

import (
	"git.intra.123u.com/rometa/romq/msg"
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"net"
	"sync"
)

type server struct {
	cfg config

	listener net.Listener

	stats

	running  bool
	shutdown bool

	accounts sync.Map
	gacc     *Account

	lock sync.RWMutex

	rwmClients sync.RWMutex
	clients    map[int64]*client
	routes     map[int64]*client
	remotes    map[string]*client

	totalClients uint64

	// LameDuck mode
	// 后端服务正在监听端口，并且可以服务请求，但是已经明确要求客户端停止发送请求。
	// 当某个请求进入跛脚鸭状态时，它会将这个状态广播给所有已经连接的客户端。
	ldm   bool
	ldmCh chan bool

	shutdownComplete chan struct{}
}

func newServer(options ...Option) *server {
	s := &server{}
	for _, opt := range options {
		opt.apply(&s.cfg)
	}

	s.clients = make(map[int64]*client)
	s.routes = make(map[int64]*client)
	s.ldmCh = make(chan bool, 1)
	s.shutdownComplete = make(chan struct{})
	s.remotes = make(map[string]*client)

	s.gacc = NewAccount(globalAccountName)
	s.registerAccount(s.gacc)

	return s
}

func (s *server) accept() {
	s.listener, _ = net.Listen("tcp", s.cfg.Addr)
	s.running = true
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
			continue
		}
		s.acceptOneConnection(conn, CLIENT)
	}
}

func (s *server) acceptOneConnection(conn net.Conn, kind ClientType) {
	nlog.Info("accept one connection %v", conn.RemoteAddr())
	id := snowflake.GetID()
	c := newAcceptClient(id, conn, s)
	c.kind = kind
	c.init()

	c.registerWithAccount(s.globalAccount())

	s.rwmClients.Lock()

	go c.run()
	s.clients[id] = c
	s.rwmClients.Unlock()

	if kind == ROUTER {
		s.addRoute(c, c.name)
	}
}

func (s *server) start() {
	nlog.Info("start server  %v %v", s.cfg.Addr, s.cfg.ClusterAddr)
	go s.accept()

	if s.cfg.ClusterAddr != "" {
		go s.startRouterListener()
	}

	if s.cfg.ConnectRouterAddr != "" {
		s.connectToRoute()
	}

	s.WaitForShutdown()
}

func (s *server) WaitForShutdown() {
	<-s.shutdownComplete
}

func (s *server) globalAccount() *Account {
	s.lock.RLock()
	defer s.lock.RUnlock()
	rs := s.gacc
	return rs
}

func (s *server) addRoute(c *client, name string) bool {
	s.lock.Lock()
	if _, ok := s.routes[c.id]; ok {
		s.lock.Unlock()
		nlog.Erro("addRoute: client %v already in routes", c.id)
		return false
	}
	if _, ok := s.remotes[name]; ok {
		s.lock.Unlock()
		nlog.Erro("addRoute: name %v already in remotes", name)
		return false
	}
	s.routes[c.id] = c
	s.remotes[name] = c
	s.lock.Unlock()

	s.sendSubsToRoute(c)

	s.forwardNewRouteInfoToKnownServers(c)
	return true
}

// 本身自己涵盖了所有subscribe的路由 ，所以只用一个 发送snapshot 过去就行了
//func (s *server) sendSubsToRoute(route *client) {
//	s.lock.Lock()
//
//	eSize := 0
//	accs := make([]*Account, 0, 32)
//	s.accounts.Range(func(key, value interface{}) bool {
//		acc := value.(*Account)
//		accs = append(accs, acc)
//		acc.rwmu.RLock()
//
//		if ns := len(acc.rm); ns > 0 {
//			eSize += ns * (len(acc.name) + 1 + 2)
//			for key := range acc.rm {
//				eSize += len(key) + 8
//			}
//		}
//
//		acc.rwmu.RUnlock()
//		return true
//	})
//
//	s.lock.Unlock()
//
//	buf := make([]byte, 0, eSize)
//	route.mu.Lock()
//	for _, a := range accs {
//		a.rwmu.RLock()
//
//		for key, value := range a.rm {
//			var subj, qn []byte
//			s := strings.Split(key, " ")
//			subj = []byte(s[0])
//
//			if len(s) > 1 {
//				qn = []byte(s[1])
//			}
//
//			sub := &subscription{
//				client:  nil,
//				subject: string(subj),
//				queue:   qn,
//				qw:      value,
//				closed:  0,
//			}
//
//			route.addRouteSubOrUnsubProtoToBuf(buf, a.name, sub, true)
//		}
//
//		a.rwmu.RUnlock()
//	}
//
//	route.mu.Unlock()
//
//	route.SendMsg(msg.MSG_SNAPSHOTSUBS, buf)
//
//	nlog.Info("send local subs to route %v  buf  %v", route.id, len(buf))
//}

func (s *server) sendSubsToRoute(route *client) {
	all := s.getAllAccountInfo()
	route.SendMsg(msg.MSG_SNAPSHOTSUBS, &msg.MsgSnapshotSubs{All: all})
}

// 获取所有Account的信息
func (s *server) getAllAccountInfo() []*msg.Accounts {
	s.lock.RLock()
	defer s.lock.RUnlock()

	accs := make([]*msg.Accounts, 0, 32)
	s.accounts.Range(func(key, value interface{}) bool {
		acc := value.(*Account)
		acc.rwmu.RLock()
		itemp := acc.getMsgAccounts()
		accs = append(accs, &itemp)
		acc.rwmu.RUnlock()
		return true
	})
	return accs
}

// 告知其他路由，有新的路由加入
func (s *server) forwardNewRouteInfoToKnownServers(route2 *client) {
	s.lock.Lock()
	for _, route := range s.routes {
		if route.id == route2.id {
			continue
		}
		route.SendMsg(msg.MSG_NEWROUTE, &msg.MsgNewRoute{Name: route2.name})
	}
	s.lock.Unlock()
}

func (s *server) removeRoute(c *client) {
	s.lock.Lock()
	delete(s.routes, c.id)
	delete(s.remotes, c.name)
	s.lock.Unlock()
}

func (s *server) snapshotSubs(snapShot *msg.MsgSnapshotSubs) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, acc := range snapShot.All {
		nacc := NewAccount(acc.Name)

		for k, v := range acc.RM {
			nacc.rm[k] = v
		}

		s.accounts.Store(acc.Name, nacc)
	}
}

func (s *server) registerAccount(account *Account) {
	s.accounts.Store(account.name, account)
}

func (s *server) updateRouteSubscriptionMap(acc *Account, sub *subscription) {
	for _, route := range s.routes {
		route.sendRemoteNewSub(sub)
	}
}

func (s *server) startRouterListener() {
	l, err := net.Listen("tcp", s.cfg.ClusterAddr)
	if err != nil {
		nlog.Erro("startRouterListener: %v", err)
		return
	}

	nlog.Info("startRouterListener: %v", s.cfg.ClusterAddr)

	for s.running {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
			continue
		}
		s.acceptOneConnection(conn, ROUTER)
	}
}

func (s *server) connectToRoute() {
	if s.cfg.ConnectRouterAddr == "" {
		return
	}

	c := newConnectClient(s.cfg.ConnectRouterAddr)
	if !c.connect() {
		nlog.Erro("connect error")
		return
	}
	c.init()
	c.srv = s
	c.register()
	go c.run()

	// 把client 加入 server
	s.lock.Lock()
	s.routes[c.id] = c
	s.lock.Unlock()
}
