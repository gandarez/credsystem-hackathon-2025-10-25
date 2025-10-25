package client

import "time"

const (
	defaultDialTimeout         = 300 * time.Millisecond
	defaultKeepAlive           = 30 * time.Second
	defaultMaxIdleConns        = 512
	defaultMaxIdleConnsPerHost = 256
	defaultIdleConnTimeout     = 30 * time.Second
	defaultTimeout             = 15 * time.Second
)

const (
	openRouterUrl = "https://openrouter.ai/api"
)
