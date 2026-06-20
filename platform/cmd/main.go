package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mirai-zen/forge/platform/internal/config"
	"github.com/mirai-zen/forge/platform/internal/handler"
	"github.com/mirai-zen/forge/platform/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "", "config file path (default: /app/configs/platform.yaml in K8s, configs/dev.yaml locally)")

func main() {
	flag.Parse()

	// 确定配置文件路径
	configPath := *configFile
	if configPath == "" {
		if _, err := os.Stat("/app/configs/platform.yaml"); err == nil {
			configPath = "/app/configs/platform.yaml" // K8s ConfigMap 挂载
			fmt.Fprintf(os.Stderr, "[config] found /app/configs/platform.yaml\n")
		} else {
			configPath = "configs/dev.yaml" // 本地开发
			fmt.Fprintf(os.Stderr, "[config] /app/configs/platform.yaml not found: %v, fallback to %s\n", err, configPath)
		}
	}

	var c config.Config
	if _, err := os.Stat(configPath); err == nil {
		conf.MustLoad(configPath, &c)
		j, _ := json.MarshalIndent(c, "", "  ")
		fmt.Fprintf(os.Stderr, "[config] go-zero loaded %s:\n%s\n", configPath, string(j))

		// 用标准 yaml 库再解析一次作对比
		raw, _ := os.ReadFile(configPath)
		var c2 config.Config
		if err := yaml.Unmarshal(raw, &c2); err != nil {
			fmt.Fprintf(os.Stderr, "[config] yaml.Unmarshal error: %v\n", err)
		} else {
			j2, _ := json.MarshalIndent(c2, "", "  ")
			fmt.Fprintf(os.Stderr, "[config] yaml.v3 loaded:\n%s\n", string(j2))
		}
	} else {
		// 容器环境无文件：从环境变量构建
		fmt.Fprintf(os.Stderr, "[config] %s not found: %v, using env\n", configPath, err)
		c = config.Config{
			RestConf: rest.RestConf{
				Host:    "0.0.0.0",
				Port:    8880,
				Timeout: 30000,
			},
		}
		c.MySQL.DataSource = getEnv("MYSQL_DSN", "")
		fmt.Fprintf(os.Stderr, "[config] env DSN=%s\n", c.MySQL.DataSource)
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
