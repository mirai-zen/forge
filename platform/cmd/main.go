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

var configFile = flag.String("f", "", "config file path (default: /app/configs/platform.yaml in K8s, configs/dev.yaml locally)")

func main() {
	flag.Parse()

	// 确定配置文件路径
	configPath := *configFile
	if configPath == "" {
		if _, err := os.Stat("/app/configs/platform.yaml"); err == nil {
			configPath = "/app/configs/platform.yaml" // K8s ConfigMap 挂载
		} else {
			configPath = "configs/dev.yaml" // 本地开发
		}
	}

	var c config.Config
	if _, err := os.Stat(configPath); err == nil {
		conf.MustLoad(configPath, &c)
	} else {
		// 容器环境无文件：从环境变量构建
		c = config.Config{
			RestConf: rest.RestConf{
				Host:    "0.0.0.0",
				Port:    8880,
				Timeout: 30000,
			},
		}
		c.MySQL.DataSource = getEnv("MYSQL_DSN", "")
	}

	// 敏感值始终优先从环境变量获取（K8s Secret 注入）
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		c.GitHub.Token = t
	}
	if o := os.Getenv("GITHUB_ORG"); o != "" {
		c.GitHub.Org = o
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterRoutes(server, ctx)

	fmt.Printf("🚀 Platform service starting at %s:%d (config: %s)\n", c.Host, c.Port, configPath)
	server.Start()
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
