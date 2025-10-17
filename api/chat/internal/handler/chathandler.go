package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"GoZero-AI/api/chat/internal/logic"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/api/chat/internal/utils"
)

func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置 SSE 头部：告诉浏览器这是一个流式响应
		setSSEHeader(w)
		flusher, _ := w.(http.Flusher)
		// 立即刷新头部，注意，由于前端请求为 post 请求，所以这里不能立即刷新头部
		// 如果前端请求为 get 请求，则可以立即刷新头部
		// flusher.Flush()

		// 2. 解析请求参数
		var req types.InterviewAppChatReq
		if err := httpx.Parse(r, &req); err != nil {
			sendSSEError(w, flusher, err.Error())
			return
		}

		// 新增：处理 PDF 文件（如果有的话）
		var pdfContent string
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()

			// 验证文件类型
			if header.Header.Get("Content-Type") != "application/pdf" {
				http.Error(w, "文件类型错误，请上传 PDF 文件", http.StatusBadRequest)
				return
			}

			// 新增：使用 mcp 服务提取 PDF 文本
			if content, err := svcCtx.PdfClient.ExtractTextFromPDF(file, header.Filename); err == nil {
				pdfContent = content
			} else {
				logx.Errorf("PDF提取失败: %v", err)
			}
		}
		// 使用 utils.CombineMessages 拼接用户消息和 PDF 内容
		req.Message = utils.CombineMessages(req.Message, pdfContent)

		// 3. 创建取消上下文
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel() // 确保资源释放

		// 4. 创建 Logic 实例，并调用业务方法
		l := logic.NewChatLogic(ctx, svcCtx)
		respChan, err := l.Chat(&req) // 注意：这里返回的是一个 channel
		if err != nil {
			sendSSEError(w, flusher, err.Error())
			return
		}

		// 5. 监听 channel，并将数据流式写入响应
		for {
			select {
			case <-ctx.Done(): // 如果客户端断开连接
				return
			case resp, ok := <-respChan: // 从 channel 读取 Logic 层传来的数据
				if !ok { // channel 已关闭
					_, err := fmt.Fprint(w, "event: end\ndata: {}\n\n")
					if err != nil {
						return
					} // 结束标记
					flusher.Flush()
					return
				}

				// 新增优化，前端流式输出时以 markdown 格式渲染内容
				// handler加个内容处理，符合前端markdown格式
				safeContent := strings.ReplaceAll(resp.Content, "\n", "\\n")
				safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")

				// 直接输出内容，不加JSON包装
				// 格式化为 SSE 规范的 "data: xxx\n\n" 格式
				_, err := fmt.Fprintf(w, "data: %s\n\n", safeContent)
				if err != nil {
					return
				}
				flusher.Flush() // 立即发送给客户端

				if resp.IsLatest {
					return
				}
			}
		}
	}
}

// setSSEHeader 设置服务器推送事件(SSE)的响应头
func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Transfer-Encoding", "chunked")
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	_, fprintf := fmt.Fprintf(w, "event: error\ndata: {\"error\":\"%s\"}\n\n", errMsg)
	if fprintf != nil {
		return
	}
	flusher.Flush()
}

// // 创建一个解析简历 PDF 的函数，用来提取简历中的文本信息
// func parseResumePDF(w http.ResponseWriter, r *http.Request, req *types.InterviewAppChatReq) (string, error) {
// 	var pdfContent string
// 	if file, header, err := r.FormFile("file"); err == nil {
// 		defer func(file multipart.File) {
// 			err := file.Close()
// 			if err != nil {
// 				logx.Errorf("关闭文件失败: %v", err)
// 				return
// 			}
// 		}(file)

// 		// 验证文件类型
// 		if header.Header.Get("Content-Type") != "application/pdf" {
// 			http.Error(w, "文件类型错误，请上传 PDF 文件", http.StatusBadRequest)
// 			return "", err
// 		}

// 		logx.Infof("检测到上传的文件简历文件了: %s", header.Filename)

// 		// 新增：使用 mcp 服务提取 PDF 文本
// 		if content, err := svcCtx.PdfClient.ExtractTextFromPDF(file, header.Filename); err == nil {
// 			pdfContent = content
// 		} else {
// 			logx.Errorf("PDF提取失败: %v", err)
// 		}

// 	}
// 	// 使用 utils.CombineMessages 拼接用户消息和 PDF 内容
// 	req.Message = utils.CombineMessages(req.Message, pdfContent)
// 	return pdfContent, nil
// }
