package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mirai-zen/forge/platform/internal/config"
	"github.com/mirai-zen/forge/platform/internal/handler"
	"github.com/mirai-zen/forge/platform/internal/model"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "", "config file path")

func main() {
	flag.Parse()

	configPath := *configFile
	if configPath == "" {
		if _, err := os.Stat("/app/configs/platform.yaml"); err == nil {
			configPath = "/app/configs/platform.yaml" // K8s ConfigMap
		} else {
			configPath = "configs/dev.yaml" // 本地开发
		}
	}

	var c config.Config
	fmt.Printf("Loading config from: %s\n", configPath)
	conf.MustLoad(configPath, &c)

	// 敏感值从 K8s Secret 注入的环境变量覆盖
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		c.GitHub.Token = t
	}
	if o := os.Getenv("GITHUB_ORG"); o != "" {
		c.GitHub.Org = o
	}

	// ────────── 可观测性初始化 ──────────

	// OpenTelemetry 链路追踪
	if c.Telemetry.Endpoint != "" {
		trace.StartAgent(c.Telemetry)
		defer trace.StopAgent()
		fmt.Printf("Trace agent: %s -> %s\n", c.Telemetry.Name, c.Telemetry.Endpoint)
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 自动迁移数据库 Schema
	ctx.DB.AutoMigrate(&model.Service{}, &model.ServiceEnv{})

	handler.RegisterRoutes(server, ctx)

	fmt.Printf("Platform starting at %s:%d\n", c.Host, c.Port)
	server.Start()
}
