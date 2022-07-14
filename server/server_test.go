package server

import (
	"git.intra.123u.com/rometa/romq/utils/snowflake"
	nlog "github.com/abc463774475/my_tool/n_log"
	"testing"
)

func TestServer_common(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	snowflake.Init(1)
	s := newServer(WithAddr(":8087"),
		WithClusterAddr(":18087"),
	)

	s.start()

	nlog.Info("end")
}
