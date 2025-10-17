package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	UniPDFLicense string // 新增 UniPDF 商业许可证密钥​
}
