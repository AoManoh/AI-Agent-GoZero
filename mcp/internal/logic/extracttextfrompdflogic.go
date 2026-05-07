// Package logic 提供MCP gRPC服务的业务逻辑实现层
// ExtractTextFromPDFLogic.go 实现PDF文档的流式处理和文本提取核心业务逻辑
//
// 主要功能:
//  1. 流式文件接收 - 处理客户端分块上传的PDF文件数据
//  2. 文件格式验证 - 确保上传文件为合法的PDF格式
//  3. 临时文件管理 - 安全地处理大文件的临时存储和清理
//  4. 文本内容提取 - 调用utils工具包完成PDF到文本的转换
//  5. 流式响应处理 - 向客户端返回提取结果或错误信息
//
// 架构定位:
//
//	本层作为GoZero框架中的Logic层，承接Handler和Utils之间的业务协调:
//	- gRPC Server -> ExtractTextFromPDFLogic -> utils.ExtractTextFromPDF
//	- 负责业务流程控制、参数验证、错误处理和资源管理
//	- 实现gRPC流式协议的具体业务逻辑
//
// 技术特性:
//   - 支持大文件的流式处理，避免内存溢出
//   - 完善的错误处理和日志记录机制
//   - 自动临时文件清理，防止磁盘空间泄露
//   - 类型安全的gRPC消息处理
//
// **后续**优化方向:
//   - 添加文件大小限制和处理进度回调
//   - 实现并发处理和任务队列机制
//   - 集成缓存机制避免重复处理相同文件
//   - 添加文件安全检查和病毒扫描
package logic

import (
	"context"
	"io"
	"os"
	"strings"

	"GoZero-AI/internal/mcpsecurity"
	"GoZero-AI/mcp/internal/svc"
	"GoZero-AI/mcp/types/mcp"
	"GoZero-AI/mcp/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ExtractTextFromPDFLogic PDF文本提取业务逻辑处理器
// 实现MCP gRPC服务中PDF文档处理的核心业务逻辑
//
// 结构设计:
//   - ctx: 请求上下文，用于取消传播和超时控制
//   - svcCtx: 服务上下文，提供配置信息和共享资源访问
//   - Logger: GoZero框架的结构化日志组件
//
// 设计模式:
//   - 采用GoZero框架的Logic层模式，实现业务逻辑与数据访问的分离
//   - 依赖注入模式，通过ServiceContext获取所需的服务组件
//   - 上下文传播模式，支持请求取消和超时处理
//
// 核心职责:
//  1. **协议处理** - 处理gRPC流式协议的复杂性
//  2. **业务流程** - 协调文件接收、验证、处理和响应的完整流程
//  3. **资源管理** - 管理临时文件的创建和清理
//  4. **错误处理** - 提供统一的错误处理和日志记录
type ExtractTextFromPDFLogic struct {
	ctx         context.Context     // 请求上下文，支持取消传播和超时控制
	svcCtx      *svc.ServiceContext // 服务上下文，提供配置和共享资源访问
	logx.Logger                     // 结构化日志组件，支持上下文关联的日志记录
}

// NewExtractTextFromPDFLogic 创建PDF文本提取业务逻辑处理器实例
// 遵循GoZero框架的工厂模式，为每个gRPC请求创建独立的逻辑处理器
//
// 设计理念:
//  1. **请求隔离** - 每个gRPC请求都有独立的Logic实例，避免并发冲突
//  2. **上下文传递** - 将gRPC请求的context传递到业务逻辑层，支持取消和超时
//  3. **依赖注入** - 通过ServiceContext注入所需的服务组件和配置
//  4. **日志关联** - 绑定请求上下文到日志系统，便于链路追踪
//
// **后续**调用时机:
//   - 在PdfProcessorServer中为每个ExtractText请求创建
//   - 实例生命周期与gRPC请求绑定，请求结束后自动回收
//   - 支持请求级别的配置覆盖和状态管理
//
// 参数说明:
//
//	ctx: gRPC请求的上下文，包含超时、取消信号和请求元数据
//	svcCtx: 服务上下文，包含配置信息、外部依赖和共享资源
//
// 返回值:
//
//	*ExtractTextFromPDFLogic: 配置完成的业务逻辑处理器实例
func NewExtractTextFromPDFLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExtractTextFromPDFLogic {
	return &ExtractTextFromPDFLogic{
		ctx:    ctx,                   // 保存请求上下文，用于后续操作的取消控制
		svcCtx: svcCtx,                // 注入服务上下文，提供业务处理所需的依赖
		Logger: logx.WithContext(ctx), // 创建带上下文的日志记录器，支持请求追踪
	}
}

// ExtractTextFromPDF 执行流式PDF文本提取的核心业务逻辑
// 实现gRPC双向流式协议，处理客户端上传的PDF文件并返回文本内容
//
// 业务流程设计:
//  1. **元数据接收** - 首先接收包含文件信息的元数据消息
//  2. **格式验证** - 验证MIME类型确保为合法的PDF文件
//  3. **临时存储** - 创建临时文件用于存储流式数据
//  4. **数据流接收** - 循环接收并写入所有数据块
//  5. **文本提取** - 调用PDF解析工具提取文本内容
//  6. **结果返回** - 封装并返回处理结果给客户端
//
// 流式协议处理:
//   - 支持大文件的分块传输，避免内存限制
//   - 使用io.EOF检测数据流结束，确保数据完整性
//   - 每个数据块都进行错误检查，及时发现传输问题
//
// 资源管理策略:
//   - 使用defer确保临时文件被正确清理
//   - 采用os.CreateTemp创建安全的临时文件
//   - 支持并发安全的文件操作
//
// 错误处理机制:
//   - 提供详细的错误日志记录，便于问题诊断
//   - 区分不同类型的错误（网络、文件、解析等）
//   - 使用SendAndClose统一处理错误响应
//
// **后续**优化可能:
//   - 添加文件大小限制和进度回调
//   - 实现文件类型的更精确检测
//   - 集成文件安全检查和病毒扫描
//   - 支持断点续传和重试机制
//
// 参数说明:
//
//	stream: gRPC流式服务器，用于接收客户端数据和返回响应
//
// 返回值:
//
//	error: 处理过程中的任何错误，包括网络、文件或解析错误
func (l *ExtractTextFromPDFLogic) ExtractTextFromPDF(stream mcp.PdfProcessor_ExtractTextFromPDFServer) error {
	if err := l.authorizeStream(stream); err != nil {
		return err
	}

	// 步骤1: 接收客户端发送的PDF文件元数据
	// 第一条消息必须包含文件的元数据信息（文件名、MIME类型等）
	metaData, err := stream.Recv()
	if err != nil {
		logx.Errorf("接收客户端发送的PDF文件元数据失败: %v", err)
		return err
	}

	// 提取并验证元数据的存在性
	// GetMetadata()返回的是指针，需要检查是否为nil
	data := metaData.GetMetadata()
	if data == nil {
		logx.Errorf("接收客户端发送的PDF文件元数据为空")
		return stream.SendAndClose(&mcp.PdfRes{Error: "缺少元数据"})
	}
	if strings.TrimSpace(data.Filename) == "" {
		logx.Errorf("接收客户端发送的PDF文件名为空")
		return stream.SendAndClose(&mcp.PdfRes{Error: "文件名不能为空"})
	}

	// 步骤2: 验证数据合法性，确保为PDF文件
	// MIME类型检查是第一道安全防线，防止恶意文件上传
	if data.MimeType != "application/pdf" {
		logx.Errorf("接收客户端发送的PDF文件类型不合法: %s", data.MimeType)
		return stream.SendAndClose(&mcp.PdfRes{Error: "文件类型不合法"})
	}

	// 步骤3: 创建临时文件，用于存储PDF文件
	// 使用os.CreateTemp创建安全的临时文件，避免文件名冲突
	tempFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		logx.Errorf("创建临时文件失败: %v", err)
		return err
	}
	// 使用defer确保资源清理，防止磁盘空间泄露
	defer os.Remove(tempFile.Name()) // 删除临时文件
	defer tempFile.Close()           // 关闭文件句柄

	// 步骤4: 处理首个数据块（如果存在）
	// 第一条消息可能同时包含元数据和数据块
	var receivedBytes int64
	if chunk := metaData.GetChunk(); chunk != nil {
		if err := l.checkUploadSize(&receivedBytes, len(chunk)); err != nil {
			return err
		}
		if _, err := tempFile.Write(chunk); err != nil {
			logx.Errorf("写入首个数据块失败: %v", err)
			return err
		}
	}

	// 步骤5: 接收后续数据块
	// 使用无限循环接收数据，直到遇到EOF或错误
	for {
		// 接收下一个数据块
		req, err := stream.Recv()
		if err == io.EOF {
			// 数据流结束，跳出循环
			break
		}
		if err != nil {
			// 网络或协议错误，记录并返回
			logx.Errorf("接收后续数据块失败: %v", err)
			return err
		}

		// 提取并写入数据块（如果存在）
		if chunk := req.GetChunk(); chunk != nil {
			if err := l.checkUploadSize(&receivedBytes, len(chunk)); err != nil {
				return err
			}
			if _, err := tempFile.Write(chunk); err != nil {
				logx.Errorf("写入后续数据块失败: %v", err)
				return err
			}
		}
	}

	// 步骤6: 提取文本，解析PDF
	// 调用本地工具函数完成PDF到文本的转换
	pdfContent, err := extractPdfText(tempFile.Name())
	if err != nil {
		logx.Errorf("提取文本失败: %v", err)
		// 返回带错误信息的响应，帮助客户端诊断问题
		return stream.SendAndClose(&mcp.PdfRes{Error: "PDF 提取文本失败" + err.Error()})
	}

	// 步骤7: 返回提取的文本
	// 使用SendAndClose关闭流并返回最终结果
	return stream.SendAndClose(&mcp.PdfRes{Content: pdfContent})
}

func (l *ExtractTextFromPDFLogic) authorizeStream(stream mcp.PdfProcessor_ExtractTextFromPDFServer) error {
	expectedToken := l.svcCtx.Config.PDF.AuthToken
	if expectedToken == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "缺少 MCP 鉴权信息")
	}
	values := md.Get(mcpsecurity.AuthTokenMetadataKey)
	if len(values) == 0 || !mcpsecurity.TokenMatches(expectedToken, values[0]) {
		return status.Error(codes.Unauthenticated, "MCP 鉴权失败")
	}

	return nil
}

func (l *ExtractTextFromPDFLogic) checkUploadSize(receivedBytes *int64, chunkSize int) error {
	if chunkSize <= 0 {
		return nil
	}

	*receivedBytes += int64(chunkSize)
	maxBytes := l.svcCtx.Config.PDFMaxUploadBytes()
	if maxBytes > 0 && *receivedBytes > maxBytes {
		return status.Error(codes.ResourceExhausted, "PDF 文件超过大小限制")
	}

	return nil
}

// extractPdfText 从本地文件路径提取PDF文本内容
// 作为ExtractTextFromPDFLogic的内部工具方法，封装文件操作和utils调用
//
// 设计目的:
//   - 将文件路径操作和流式处理隔离
//   - 提供统一的文件访问接口，方便单元测试
//   - 隐藏utils包的调用细节，保持业务逻辑的清晰
//   - 支持未来扩展附加的文件处理逻辑
//
// 技术实现:
//   - 使用os.Open安全地打开文件，自动处理权限检查
//   - 采用defer模式确保文件句柄被正确关闭
//   - 直接返回utils.ExtractTextFromPDF的结果，保持错误信息的完整性
//
// **后续**优化可能:
//   - 添加文件大小检查和安全验证
//   - 集成缓存机制避免重复解析
//   - 添加进度回调和取消支持
//   - 支持更多文档格式的统一处理
//
// 参数说明:
//
//	filePath: 本地PDF文件的绝对路径，由临时文件系统生成
//
// 返回值:
//
//	string: 提取的纯文本内容，包含所有页面的文本信息
//	error: 文件访问或PDF解析过程中的任何错误
func extractPdfText(filePath string) (string, error) {
	// 打开本地PDF文件，进行权限和存在性检查
	file, err := os.Open(filePath)
	if err != nil {
		// 文件打开失败，可能原因:权限不足、文件不存在、路径错误等
		return "", err
	}
	// 使用defer确保文件句柄在函数退出时被关闭
	defer file.Close()

	// 调用utils包的PDF解析工具，完成文本提取
	// 返回的错误将直接传递给上层调用者
	return utils.ExtractTextFromPDF(file)
}
