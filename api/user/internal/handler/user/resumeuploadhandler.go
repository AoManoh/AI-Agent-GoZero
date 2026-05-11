package user

import (
	"errors"
	"net/http"

	logic "GoZero-AI/api/user/internal/logic/user"
	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/pdfgrpc"
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
			httpx.ErrorCtx(r.Context(), w, normalizeResumeUploadError(err))
			return
		}

		content, err := svcCtx.PdfClient.ExtractTextFromPDF(r.Context(), file, header.Filename)
		if err != nil {
			if pdfgrpc.IsUploadTooLarge(err) {
				httpx.ErrorCtx(r.Context(), w, normalizeResumeUploadError(pdfupload.ErrUploadLarge))
				return
			}
			httpx.ErrorCtx(r.Context(), w, statuserr.Coded(http.StatusBadGateway, "parser_failed", "PDF 文本提取失败，请稍后重试"))
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

func normalizeResumeUploadError(err error) error {
	switch {
	case errors.Is(err, pdfupload.ErrUploadLarge):
		return statuserr.Coded(http.StatusRequestEntityTooLarge, "file_too_large", err.Error())
	case errors.Is(err, pdfupload.ErrInvalidPDF):
		return statuserr.Coded(http.StatusBadRequest, "invalid_pdf", err.Error())
	case errors.Is(err, pdfupload.ErrEmptyPDF):
		return statuserr.Coded(http.StatusBadRequest, "empty_text", err.Error())
	default:
		return err
	}
}
