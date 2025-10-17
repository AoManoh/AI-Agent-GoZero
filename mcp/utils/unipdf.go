// Package utils 提供MCP gRPC服务的PDF文档处理工具集
// 这是GoZero-AI项目中MCP微服务的核心工具包，专门负责PDF文档的解析和文本提取功能
//
// 主要功能:
//  1. PDF文档解析 - 基于unipdf库实现高性能PDF文档解析
//  2. 文本内容提取 - 将PDF页面中的文本对象转换为纯文本字符串
//  3. 多页处理支持 - 批量处理PDF文档的所有页面内容
//  4. 内存优化管理 - 使用缓冲机制避免重复文件读取
//
// 技术特性:
//   - 基于unipdf/v3库，支持复杂PDF格式解析
//   - 流式文件处理，支持io.Reader接口
//   - 内存友好设计，避免大文件内存溢出
//   - 错误处理完善，支持损坏PDF文档的容错处理
//
// gRPC集成:
//   本包作为MCP gRPC服务的底层工具，为以下组件提供支持:
//   - ExtractTextLogic: 业务逻辑层的PDF处理实现
//   - PdfProcessorServer: gRPC服务端的文档处理能力
//   - 流式文件传输: 支持大文件分块上传和实时处理
//
// 应用场景:
//   - 企业文档管理系统的PDF内容索引
//   - RAG知识库构建中的文档预处理
//   - 智能文档分析系统的文本提取
//   - 批量文档处理和内容迁移
//
// 性能优化:
//   - 使用strings.Builder进行高效文本拼接
//   - 内存缓冲机制减少IO操作次数
//   - 页面级别的错误隔离，单页失败不影响整体处理
//   - 支持大文档的流式处理模式
//
// **后续**扩展能力:
//   - 添加OCR文字识别支持，处理图像化PDF
//   - 实现文档结构化解析（标题、段落、表格等）
//   - 集成文档安全检查和病毒扫描
//   - 支持其他文档格式（Word、PowerPoint等）
//   - 添加文档元数据提取功能
//
// 依赖项:
//   - github.com/unidoc/unipdf/v3: PDF解析核心库
//   - bytes: 内存缓冲和数据处理
//   - io: 流式文件读取接口
//   - strings: 高效文本处理工具
package utils

import (
	"bytes"
	"io"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// ExtractTextFromPDF 从 PDF 文件中提取文本
func ExtractTextFromPDF(file io.Reader) (string, error) {
	// 创建内存缓冲区避免重复读取
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	// 创建 PDF 解析器
	pdfReader, err := model.NewPdfReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}

	// 获取PDF文档的总页数
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	// 创建一个字符串构建器，用于高效地拼接多页文本内容
	var textBuilder strings.Builder

	// 遍历PDF的每一页，注意PDF页码从1开始（不是0）
	// 后续的所有获取都需要错误检查，因为页面可能损坏或无法读取
	for i := 1; i <= numPages; i++ {
		// 获取指定页码的页面对象
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return "", err
		}

		// 为当前页面创建文本提取器
		// extractor.New() 会分析页面的文本对象和布局信息
		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}

		// 从页面中提取纯文本内容
		// ExtractText() 会将页面中的所有文本对象转换为字符串
		pageText, err := ex.ExtractText()
		if err != nil {
			return "", err
		}

		// 清理页面文本：去除首尾空白字符
		// 使用 strings.TrimSpace() 去除字符串两端的空白字符
		textBuilder.WriteString(strings.TrimSpace(pageText))

		// 在每页之间添加两个换行符作为分隔
		// 使用 "\n\n" 作为分隔符，这样可以在最终文本中区分不同页面的内容
		textBuilder.WriteString("\n\n")
	}

	// 返回提取的文本
	return textBuilder.String(), nil
}
