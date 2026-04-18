package user

import (
	"net/http"

	logic "GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/pdfupload"
	"GoZero-AI/internal/statuserr"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ResumeUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResumeUploadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, pdfupload.RequiredFormFileError(err))
			return
		}
		defer file.Close()

		if err := pdfupload.ValidatePDFUpload(file, header); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		content, err := svcCtx.PdfClient.ExtractTextFromPDF(r.Context(), file, header.Filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, statuserr.New(http.StatusBadGateway, "PDF 文本提取失败，请稍后重试"))
			return
		}

		l := logic.NewResumeUploadLogic(r.Context(), svcCtx)
		resp, err := l.ResumeUpload(&req, header.Filename, content)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
