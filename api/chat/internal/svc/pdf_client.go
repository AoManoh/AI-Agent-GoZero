// Package svc 提供API服务层的gRPC客户端封装和服务依赖管理
// pdf_client.go 实现与MCP gRPC服务通信的PDF文档处理客户端
//
// 主要功能:
//  1. gRPC客户端封装 - 提供高级别的PDF处理接口，隐藏底层gRPC通信细节
//  2. 流式数据传输 - 支持大文件的分块上传和流式处理机制
//  3. 错误处理适配 - 将gRPC错误转换为API层可理解的错误格式
//  4. 连接管理优化 - 使用连接池和非阻塞模式提高性能
//  5. 元数据处理协调 - 管理文件元信息和数据块的正确传输顺序
//
// 架构定位:
//
//	本文件作为API服务的gRPC客户端适配器，在微服务架构中承担以下角色:
//	- API HTTP Handler -> PdfClient -> MCP gRPC Service
//	- 实现HTTP multipart/form-data到gRPC流式协议的转换
//	- 提供统一的PDF处理接口给上层业务逻辑使用
//	- 隔离API服务与MCP服务的通信复杂性
//
// 技术特性:
//   - 基于GoZero zrpc客户端框架，支持服务发现和负载均衡
//   - 采用流式gRPC协议，支持大文件传输而不占用过多内存
//   - 非阻塞连接模式，提高并发处理能力
//   - 完善的错误处理和日志记录机制
//   - 支持连接复用和自动重连
//
// 使用场景:
//   - RAG知识库构建中的PDF文档预处理
//   - 企业文档管理系统的内容提取
//   - 智能文档分析平台的文本抽取
//   - 批量文档处理任务的客户端接口
//
// **后续**扩展能力:
//   - 添加重试机制和熔断器模式
//   - 实现连接池的动态管理和监控
//   - 支持多种文档格式的统一处理接口
//   - 集成缓存机制避免重复处理
//   - 添加处理进度回调和取消机制
//
// 依赖项:
//   - GoZero-AI/mcp/types/mcp: MCP gRPC服务的接口定义
//   - github.com/zeromicro/go-zero/zrpc: GoZero的gRPC客户端框架
//   - mime/multipart: HTTP文件上传的标准处理
package svc

import (
	"context"
	"mime/multipart"

	"GoZero-AI/internal/pdfgrpc"
	"GoZero-AI/mcp/types/mcp"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// PdfClient PDF文档处理gRPC客户端封装器
// 作为API服务层与MCP gRPC服务之间的适配器，提供统一的PDF处理能力
//
// 设计模式:
//   - **适配器模式**: 将底层gRPC复杂性隐藏，提供简洁的API接口
//   - **代理模式**: 代理访问MCP服务，实现透明的远程调用
//   - **单例模式**: 通常在ServiceContext中作为单例使用，共享连接资源
//
// 核心责任:
//  1. **协议转换** - HTTP multipart文件上gRPC流式数据的转换
//  2. **连接管理** - 管理与MCP服务的gRPC连接生命周期
//  3. **错误处理** - 将gRPC错误转换为业务层可理解的错误格式
//  4. **资源优化** - 合理管理内存和网络资源的使用
//
// 结构字段:
//   - Client: MCP gRPC服务的客户端实例，负责实际的远程调用
type PdfClient struct {
	Client mcp.PdfProcessorClient // MCP gRPC服务的客户端实例，由GoZero zrpc框架管理
}

// NewPdfClient 创建PDF文档处理gRPC客户端实例
// 使用GoZero zrpc框架初始化与MCP服务的gRPC连接
//
// 初始化配置:
//  1. **服务端点**: 通过endPoint参数指定MCP服务的地址
//  2. **非阻塞模式**: 设置NonBlock为true，提高并发处理能力
//  3. **连接复用**: zrpc框架自动管理连接池和连接复用
//  4. **错误处理**: 使用MustNewClient确保初始化必须成功
//
// 技术实现:
//   - 采用GoZero的zrpc客户端框架，自带服务发现和负载均衡
//   - 支持多个端点配置，实现高可用性
//   - 自动处理连接断开和重连机制
//   - 集成链路追踪和监控指标
//
// **后续**使用模式:
//   - 通常在ServiceContext中作为单例初始化
//   - 多个API请求共享同一个客户端实例
//   - 支持请求级别的超时和取消控制
//
// 参数说明:
//
//	endPoint: MCP gRPC服务的网络地址，格式为"host:port"
//
// 返回值:
//
//	*PdfClient: 配置完成的PDF客户端实例，可用于后续的PDF处理调用
func NewPdfClient(endPoint string) *PdfClient {
	// 创建zrpc客户端连接，使用MustNewClient确保初始化成功
	// NonBlock: true 设置为非阻塞模式，提高并发性能
	conn := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{endPoint}, // 支持多个端点实现高可用性
		NonBlock:  true,               // 非阻塞连接模式，提高响应性能
	},
		// 放大发送/接收消息限制
		zrpc.WithDialOption(grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(50*1024*1024),
			grpc.MaxCallRecvMsgSize(50*1024*1024),
		)),
	)

	// 返回配置完成的PdfClient实例
	return &PdfClient{
		// 使用生成的gRPC客户端创建MCP服务的PdfProcessor客户端
		Client: mcp.NewPdfProcessorClient(conn.Conn()),
	}
}

// ExtractTextFromPDF 执行PDF文本提取的高级业务接口
// 将HTTP multipart文件上传转换为gRPC流式协议，调用MCP服务完成PDF解析
//
// 业务流程设计:
//  1. **gRPC流初始化** - 创建与MCP服务的双向流连接
//  2. **元数据传输** - 首先发送文件的元信息（文件名、MIME类型）
//  3. **文件数据传输** - 将整个文件作为单个数据块发送
//  4. **响应接收** - 接收MCP服务返回的解析结果
//  5. **错误处理** - 处理各种可能的错误情况并返回统一格式
//
// 协议转换实现:
//   - 将HTTP multipart.File转换为gRPC流式数据
//   - 使用oneof联合类型实现元数据和数据块的分开发送
//   - 采用CloseAndRecv模式实现与服务端的同步通信
//
// 性能优化策略:
//   - 使用io.ReadAll一次性读取文件，减少IO操作次数
//   - 采用defer模式确保流连接的正确关闭
//   - 早期错误检测，避免不必要的数据传输
//
// 错误处理机制:
//   - gRPC连接错误: 网络故障或服务不可用
//   - 文件读取错误: multipart文件损坏或内存不足
//   - 业务逻辑错误: MCP服务返回的PDF解析失败
//   - 所有错误都会记录到日志并返回给调用者
//
// **后续**优化方向:
//   - 实现真正的流式传输，支持大文件分块处理
//   - 添加文件大小限制和类型校验
//   - 集成超时和取消机制
//   - 添加重试和熔断器支持
//   - 实现处理进度回调
//
// 参数说明:
//
//	file: HTTP multipart文件上传的文件对象，包含PDF文件数据
//	filename: 原始文件名，用于日志记录和错误追踪
//
// 返回值:
//
//	string: 提取的PDF文本内容，包含所有页面的文本信息
//	error: 处理过程中的任何错误，包括网络、文件或解析错误
func (c *PdfClient) ExtractTextFromPDF(ctx context.Context, file multipart.File, filename string) (string, error) {
	content, err := pdfgrpc.ExtractText(ctx, func(ctx context.Context) (pdfgrpc.ClientStream, error) {
		stream, err := c.Client.ExtractTextFromPDF(ctx)
		if err != nil {
			logx.Errorf("创建 gRPC 流式客户端失败: %v", err)
			return nil, err
		}
		return stream, nil
	}, file, filename)
	if err != nil {
		logx.Errorf("PDF 文本提取失败: %v", err)
		return "", err
	}

	return content, nil
}
