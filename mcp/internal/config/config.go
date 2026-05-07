package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	UniPDFLicense string // 新增 UniPDF 商业许可证密钥​
	PDF           PDFConfig
}

type PDFConfig struct {
	AuthToken      string `json:",optional"`
	MaxUploadBytes int64  `json:",optional"`
}

func (c Config) PDFMaxUploadBytes() int64 {
	if c.PDF.MaxUploadBytes <= 0 {
		return 50 * 1024 * 1024
	}
	return c.PDF.MaxUploadBytes
}
