// Package handler 提供HTTP请求处理器实现
// knowledge_upload_logic.go 实现RAG知识库的PDF文档上传处理功能
//
// 主要功能:
//  1. PDF文件上传接收与验证 - 处理multipart/form-data请求，验证文件类型
//  2. PDF文本内容提取 - 调用unipdf库解析PDF文档获取文本内容
//  3. 知识库存储协调 - 调用Logic层实现文档分块和向量化存储
//  4. HTTP响应处理 - 返回标准化的JSON响应给前端
//
// 技术特性:
//   - 支持多媒体文件上传(multipart/form-data)
//   - 严格的文件类型验证(仅支持PDF)
//   - 统一的错误处理和响应格式
//   - RESTful API设计模式
//
// 业务流程:
//
//	用户上传PDF -> Handler验证 -> 文本提取 -> Logic分块处理 -> VectorStore存储 -> 响应结果
//
// 应用场景:
//   - 企业知识库构建(上传技术文档、手册等)
//   - AI客服知识源建设(上传FAQ、产品说明等)
//   - 智能问答系统的知识输入(研究报告、论文等)
//   - RAG增强型AI的外部知识注入
package handler

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"GoZero-AI/api/chat/internal/logic"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
)

// KnowledgeUploadHandler 知识库文档上传处理器
// 实现RAG系统的核心功能之一：外部知识导入，为AI提供专业领域知识支持
//
// 功能职责:
//  1. **文件接收与验证** - 处理前端上传的PDF文件，验证格式合法性
//  2. **内容提取处理** - 调用PDF解析工具提取文档中的文本内容
//  3. **业务逻辑委托** - 将提取的内容传递给Logic层进行分块和向量化处理
//  4. **响应状态管理** - 向前端返回处理结果和成功/失败状态
//
// 技术实现:
//   - HTTP multipart/form-data文件上传处理
//   - Content-Type MIME类型验证确保文件安全
//   - 统一的错误处理机制，提供用户友好的错误信息
//   - RESTful API设计，符合HTTP标准
//
// 业务价值:
//   - **知识库扩展**: 允许用户上传专业文档丰富AI知识源
//   - **内容准确性**: 基于权威文档的知识比训练数据更准确和及时
//   - **定制化支持**: 企业可上传内部文档实现定制化AI服务
//   - **知识更新**: 支持知识库的持续更新和扩展
//
// 调用链路:
//
//	前端上传 -> KnowledgeUploadHandler -> ExtractTextFromPDF -> KnowledgeUploadLogic -> VectorStore
//
// **后续**扩展点:
//   - 支持更多文件格式(Word、TXT、Markdown等)
//   - 添加文件大小限制和批量上传
//   - 实现文档预处理(去重、格式化等)
//   - 增加上传进度监控和异步处理
//
// 参数说明:
//
//	svcCtx: 服务上下文，包含配置信息、依赖组件和业务逻辑实例
//
// 返回值:
//
//	http.HandlerFunc: 符合Go-Zero框架规范的HTTP处理函数
func KnowledgeUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 步骤1: 解析multipart/form-data请求，获取上传的文件
		// r.FormFile("file")解析名为"file"的表单字段
		// file: 文件内容流，header: 文件元信息(文件名、大小、类型等)
		file, header, err := r.FormFile("file")
		if err != nil {
			// 文件上传失败，可能原因：
			// - 请求格式不正确(非multipart/form-data)
			// - 字段名不匹配(前端使用的字段名不是"file")
			// - 文件大小超出服务器限制
			httpx.Error(w, err)
			return
		}
		defer file.Close() // 确保文件流及时关闭，避免资源泄露

		// 步骤2: 验证文件类型，确保上传的是PDF文件
		// 通过HTTP header中的Content-Type进行MIME类型检查
		// 这是第一道安全防线，防止恶意文件上传
		if header.Header.Get("Content-Type") != "application/pdf" {
			// 返回用户友好的错误信息，指导用户上传正确格式
			httpx.Error(w, errors.New("仅支持 PDF 文件！"))
			return
		}

		// 步骤3更新: 调用 mcp 微服务提取 PDF 文本
		// ExtractTextFromPDF使用 mcp 解析PDF文档
		// 提取出的content是纯文本，去除了格式和图片信息
		content, err := svcCtx.PdfClient.ExtractTextFromPDF(file, header.Filename)
		if err != nil {
			logx.Errorf("PDF文本提取失败: %v", err)
		}

		// 步骤4: 提取文档标题信息
		// 使用文件名作为知识文档的标题，便于后续检索结果的展示
		// 标题将与每个知识块关联，帮助用户理解知识来源
		title := header.Filename

		// 步骤5: 委托业务逻辑层处理文档存储
		// 创建KnowledgeUploadLogic实例，传入请求上下文和服务依赖
		l := logic.NewKnowledgeUploadLogic(r.Context(), svcCtx)

		// 调用业务逻辑处理文档分块、向量化和存储
		// KnowledgeUploadReq包装提取的标题和内容
		resp, err := l.KnowledgeUpload(&types.KnowledgeUploadReq{
			Title:   title,   // PDF文件名，用作文档标题
			Content: content, // 提取的文本内容，将被分块处理
		})

		// 步骤6: 处理业务逻辑响应并返回给前端
		if err != nil {
			// 业务处理失败，可能原因：
			// - 文档分块处理异常
			// - 向量化API调用失败
			// - 数据库存储错误
			httpx.Error(w, err)
		} else {
			// 处理成功，返回结构化的JSON响应
			// resp包含处理结果信息(成功消息、分块数量等)
			httpx.OkJson(w, resp)
		}
	}
}
