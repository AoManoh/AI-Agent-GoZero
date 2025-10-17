// Package utils 这个文件是 unipdf 的封装，用于处理 PDF 文件
// 封装了 PDF 文件的解析和文本提取功能
package utils

// CombineMessages 合并消息，简单拼接用户消息和 PDF 内容，但是要注意别超过最大 token
func CombineMessages(userMessage string, pdfContent string) string {
	const maxTokens = 2047

	// 如果 PDF 内容为空，直接返回用户消息
	if pdfContent == "" {
		return userMessage
	}

	// 检查 PDF 内容是否超过最大 token
	if len([]rune(pdfContent)) > maxTokens {
		return userMessage + "\n[系统提示]PDF 内容超过文本上限！"
	}

	// 正常拼接内容
	return userMessage + "\n[系统提示]PDF 内容：" + pdfContent + "[PDF内容结束]"
}

// SplitText 将长文本分割为指定大小的文本块
// 用于 RAG 知识库系统，将大文档（如 PDF）切分为可处理的小块进行向量化存储
func SplitText(text string, maxChunkSize int) []string {
	var chunks []string

	if text == "" {
		return chunks
	}

	// 转换为 rune 数组以正确处理多字节字符（如中文、emoji 等）
	// 使用 rune 而不是 byte 确保字符完整性，避免截断多字节字符
	// 八股小知识：rune 和 byte 的区别？
	runes := []rune(text)
	totalLength := len(runes)

	// 按指定大小循环分割文本
	for i := 0; i < totalLength; i += maxChunkSize {
		// 计算当前块的结束位置
		end := i + maxChunkSize

		// 边界检查：确保不超出文本总长度
		if end > totalLength {
			end = totalLength
		}

		// 将当前块转换为字符串并添加到结果数组
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// TruncateText 截断文本到指定长度
// TruncateText 截断文本到指定长度并添加省略号
// 用于 RAG 系统中控制注入到 AI 上下文的知识内容长度，防止超出模型 token 限制
// 主要应用场景:
//  1. 知识检索结果注入对话上下文时的长度控制
//  2. 配合 Knowledge.MaxContextLength 配置项使用
//  3. 确保 AI 对话的响应速度和质量
//
// 参数:
//
//	text: 待截断的原始文本内容
//	maxLength: 最大允许的字符长度（基于 rune 计算，支持多字节字符）
//
// 返回:
//
//	string: 截断后的文本，如果超长则添加 "..." 后缀
func TruncateText(text string, maxLength int) string {
	// 转换为 rune 数组以正确处理多字节字符（如中文、emoji 等）
	// 使用 rune 而不是 byte 确保字符完整性，避免截断多字节字符
	runes := []rune(text)

	// 如果文本长度未超过限制，直接返回原文本
	if len(runes) <= maxLength {
		return text
	}

	// 截断文本并添加省略号，提示用户内容被截断
	// 省略号帮助用户理解这是部分内容，提高用户体验
	return string(runes[:maxLength]) + "..."
}
