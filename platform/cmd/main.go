package main

import (
	"flag"
	"fmt"

	"github.com/mirai-zen/forge/platform/internal/config"
	"github.com/mirai-zen/forge/platform/internal/handler"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "configs/platform.yaml", "config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterRoutes(server, ctx)

	fmt.Printf("🚀 Platform service starting at %s:%d\n", c.Host, c.Port)
	server.Start()
}
