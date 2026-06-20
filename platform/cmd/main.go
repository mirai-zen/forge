package main

import (
	"flag"
	"fmt"
	"os"

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
	if _, err := os.Stat(*configFile); err == nil {
		conf.MustLoad(*configFile, &c)
	} else {
		// 容器环境：从环境变量读取
		c = config.Config{
			RestConf: rest.RestConf{
				Host:    "0.0.0.0",
				Port:    8880,
				Timeout: 30000,
			},
		}
		c.MySQL.DataSource = getEnv("MYSQL_DSN", "root:123456@tcp(mysql:3306)/forge_platform?charset=utf8mb4&parseTime=True&loc=Local")
		c.GitHub.Token = getEnv("GITHUB_TOKEN", "")
		c.GitHub.Org = getEnv("GITHUB_ORG", "mirai-zen")
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterRoutes(server, ctx)

	fmt.Printf("🚀 Platform service starting at %s:%d\n", c.Host, c.Port)
	server.Start()
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
