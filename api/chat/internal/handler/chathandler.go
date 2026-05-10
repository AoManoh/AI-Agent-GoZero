package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	chatAuth "GoZero-AI/api/chat/internal/auth"
	"GoZero-AI/api/chat/internal/logic"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/api/chat/internal/utils"
	"GoZero-AI/internal/pdfgrpc"
	"GoZero-AI/internal/pdfupload"
	"GoZero-AI/internal/statuserr"
)

func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if accessToken := bearerTokenFromHeader(r.Header.Get("Authorization")); accessToken != "" {
			userID, err := chatAuth.ParseAccessTokenUserID(svcCtx.Config.Auth.AccessSecret, accessToken)
			if err != nil {
				httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, map[string]any{
					"message": "access token 无效或已过期",
				})
				return
			}
			ctx = chatAuth.WithUserID(ctx, userID)
		}

		// 1. 解析请求参数
		var req types.InterviewAppChatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		// 新增：处理 PDF 文件（如果有的话）
		var pdfContent string
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()

			if err := pdfupload.ValidatePDFUpload(file, header); err != nil {
				httpx.ErrorCtx(ctx, w, err)
				return
			}

			// 新增：使用 mcp 服务提取 PDF 文本
			if content, err := svcCtx.PdfClient.ExtractTextFromPDF(ctx, file, header.Filename); err == nil {
				if strings.TrimSpace(content) == "" {
					httpx.ErrorCtx(ctx, w, errors.New("PDF 未提取到有效文本内容"))
					return
				}
				pdfContent = content
			} else {
				logx.Errorf("PDF提取失败: %v", err)
				if pdfgrpc.IsUploadTooLarge(err) {
					httpx.ErrorCtx(ctx, w, pdfupload.ErrUploadLarge)
					return
				}
				httpx.ErrorCtx(ctx, w, statuserr.New(http.StatusBadGateway, "PDF 文本提取失败，请稍后重试"))
				return
			}
		} else if formErr := pdfupload.OptionalFormFileError(err); formErr != nil {
			httpx.ErrorCtx(ctx, w, formErr)
			return
		}
		// 使用 utils.CombineMessages 拼接用户消息和 PDF 内容
		req.Message = utils.CombineMessages(req.Message, pdfContent)

		// 2. 创建取消上下文
		ctx, cancel := context.WithCancel(ctx)
		defer cancel() // 确保资源释放

		// 3. 创建 Logic 实例，并调用业务方法
		l := logic.NewChatLogic(ctx, svcCtx)
		respChan, err := l.Chat(&req) // 注意：这里返回的是一个 channel
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		// 4. 仅在真正进入流式阶段后再切换到 SSE 响应
		setSSEHeader(w)
		flusher, _ := w.(http.Flusher)

		// 5. 监听 channel，并将数据流式写入响应
		for {
			select {
			case <-ctx.Done(): // 如果客户端断开连接
				return
			case resp, ok := <-respChan: // 从 channel 读取 Logic 层传来的数据
				if !ok { // channel 已关闭
					if !sendSSEDone(w, flusher) {
						return
					}
					return
				}

				if resp.IsLatest {
					if resp.Content != "" && !sendSSEData(w, flusher, resp.Content) {
						return
					}
					if !sendSSEDone(w, flusher) {
						return
					}
					return
				}

				if resp.Event != "" {
					if !sendSSEEvent(w, flusher, resp.Event, resp.Content) {
						return
					}
					continue
				}

				if !sendSSEData(w, flusher, resp.Content) {
					return
				}
			}
		}
	}
}

func bearerTokenFromHeader(headerValue string) string {
	const prefix = "bearer "

	value := strings.TrimSpace(headerValue)
	if len(value) < len(prefix) || strings.ToLower(value[:len(prefix)]) != prefix {
		return ""
	}

	return strings.TrimSpace(value[len(prefix):])
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

func sendSSEData(w http.ResponseWriter, flusher http.Flusher, content string) bool {
	safeContent := strings.ReplaceAll(content, "\n", "\\n")
	safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")

	if _, err := fmt.Fprintf(w, "data: %s\n\n", safeContent); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event, content string) bool {
	safeEvent := strings.ReplaceAll(strings.TrimSpace(event), "\n", "")
	safeEvent = strings.ReplaceAll(safeEvent, "\r", "")
	if safeEvent == "" {
		return true
	}
	safeContent := strings.ReplaceAll(content, "\n", "\\n")
	safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", safeEvent, safeContent); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

func sendSSEDone(w http.ResponseWriter, flusher http.Flusher) bool {
	if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
		return false
	}
	flusher.Flush()
	return true
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
