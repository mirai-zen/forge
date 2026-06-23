package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Upstream UpstreamConfig
}

type UpstreamConfig struct {
	Platform string // http://platform-dev.forge-dev:8880
	User     string // http://user-dev.forge-dev:8881
}
