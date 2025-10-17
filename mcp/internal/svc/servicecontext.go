package svc

import (
	"GoZero-AI/mcp/internal/config"
	"fmt"

	"github.com/unidoc/unipdf/v3/common/license"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 新增：初始化简历解析器
	// 设置 UniPDF key
	err := license.SetMeteredKey(c.UniPDFLicense)
	if err != nil {
		fmt.Printf("设置 UniPDF 许可证失败: %v\n", err)
		// 注意：如果没有授权，UniPDF 会添加水印
	}

	return &ServiceContext{
		Config: c,
	}
}
