package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MySQL struct {
		DataSource string `json:"DataSource"`
	} `json:"MySQL"`
	GitHub struct {
		Token string `json:"Token"`
		Org   string `json:"Org,default=mirai-zen"`
	} `json:"GitHub"`
}
