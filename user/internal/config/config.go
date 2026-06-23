// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MySQL MysqlConfig
	JWT   JwtConfig
}

type MysqlConfig struct {
	DataSource string
}

type JwtConfig struct {
	Secret string
	Expire int64 // seconds
}
