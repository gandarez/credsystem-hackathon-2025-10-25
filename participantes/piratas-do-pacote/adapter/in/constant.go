package in

import "time"

const (
	defaultServerName         = "intention-server"
	defaultReadTimeout        = 2 * time.Second
	defaultWriteTimeout       = 2 * time.Second
	defaultReadBufferSize     = 4096
	defaultWriteBufferSize    = 4096
	defaultMaxRequestBodySize = 1 << 20
)
