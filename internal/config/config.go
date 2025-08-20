package config

import (
	"gitee.com/link234/newbee-backend-common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DatabaseConf config.DatabaseConf
	RedisConf    config.RedisConf
}
