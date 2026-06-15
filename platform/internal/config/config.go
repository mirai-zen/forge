package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MySQL struct {
		DataSource string
	}
	GitHub struct {
		Token string
		Org   string `json:",default=mirai-zen"`
	}
}
