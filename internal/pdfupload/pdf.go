package pdfupload

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"GoZero-AI/internal/statuserr"
)

var (
	ErrMissingPDF  = statuserr.New(http.StatusBadRequest, "请上传 PDF 文件")
	ErrEmptyPDF    = statuserr.New(http.StatusBadRequest, "PDF 文件为空")
	ErrInvalidPDF  = statuserr.New(http.StatusBadRequest, "仅支持 PDF 文件")
	ErrUploadParse = statuserr.New(http.StatusBadRequest, "上传文件解析失败")
	ErrUploadLarge = statuserr.New(http.StatusRequestEntityTooLarge, "上传文件过大")
)

func OptionalFormFileError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, http.ErrMissingFile), errors.Is(err, http.ErrNotMultipart):
		return nil
	case errors.Is(err, multipart.ErrMessageTooLarge):
		return ErrUploadLarge
	default:
		return ErrUploadParse
	}
}

func RequiredFormFileError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, http.ErrMissingFile), errors.Is(err, http.ErrNotMultipart):
		return ErrMissingPDF
	case errors.Is(err, multipart.ErrMessageTooLarge):
		return ErrUploadLarge
	default:
		return ErrUploadParse
	}
}

// ValidatePDFUpload validates the uploaded file using content sniffing, while
// keeping the file cursor rewound for downstream PDF extraction.
func ValidatePDFUpload(file multipart.File, header *multipart.FileHeader) error {
	if file == nil || header == nil {
		return ErrMissingPDF
	}

	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return seekErr
	}

	if header.Size == 0 || n == 0 {
		return ErrEmptyPDF
	}

	detectedContentType := http.DetectContentType(head[:n])

	if bytes.HasPrefix(head[:n], []byte("%PDF-")) || detectedContentType == "application/pdf" {
		return nil
	}

	return ErrInvalidPDF
}
