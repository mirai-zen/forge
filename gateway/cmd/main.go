package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mirai-zen/forge/gateway/internal/config"
	"github.com/mirai-zen/forge/gateway/internal/middleware"
	"github.com/mirai-zen/forge/gateway/internal/proxy"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(logx.LogConf{
		ServiceName: "gateway",
		Mode:        "console",
		Level:       "info",
	})

	// 反向代理 handler
	proxyHandler := proxy.NewProxyHandler(c.Upstream.Platform, c.Upstream.User)

	// JWT 中间件包裹
	jwtMw := middleware.JwtAuth(c.Upstream.User)
	handler := jwtMw(proxyHandler)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	logx.Infof("Gateway starting at %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logx.Errorf("gateway failed: %v", err)
		panic(err)
	}
}
