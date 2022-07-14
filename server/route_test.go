package server

import (
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"testing"
)

func TestServer_route(t *testing.T) {
	snowflake.Init(1)
	s := newServer(WithAddr(":8088"),
		WithClusterAddr(":18088"),
		WithConnectRouterAddr("127.0.0.1:18087"),
	)

	s.start()

	nlog.Info("end")
}
