package pdfupload

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"
)

type testMultipartFile struct {
	*bytes.Reader
}

func (f *testMultipartFile) Close() error {
	return nil
}

func newFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Size:     size,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
	}
}

func TestValidatePDFRejectsForgedHeader(t *testing.T) {
	payload := []byte("this is not a pdf file")
	file := &testMultipartFile{Reader: bytes.NewReader(payload)}
	header := newFileHeader("resume.pdf", "application/pdf", int64(len(payload)))

	err := ValidatePDFUpload(file, header)
	if !errorsIs(err, ErrInvalidPDF) {
		t.Fatalf("ValidatePDFUpload() error = %v, want ErrInvalidPDF", err)
	}
}

func TestValidatePDFRejectsEmptyFile(t *testing.T) {
	file := &testMultipartFile{Reader: bytes.NewReader(nil)}
	header := newFileHeader("resume.pdf", "application/pdf", 0)

	err := ValidatePDFUpload(file, header)
	if !errorsIs(err, ErrEmptyPDF) {
		t.Fatalf("ValidatePDFUpload() error = %v, want ErrEmptyPDF", err)
	}
}

func TestValidatePDFAcceptsValidPDFAndRewindsReader(t *testing.T) {
	payload := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF")
	file := &testMultipartFile{Reader: bytes.NewReader(payload)}
	header := newFileHeader("resume.pdf", "application/pdf", int64(len(payload)))

	if err := ValidatePDFUpload(file, header); err != nil {
		t.Fatalf("ValidatePDFUpload() error = %v, want nil", err)
	}

	rewound, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !strings.HasPrefix(string(rewound), "%PDF-1.4") {
		t.Fatalf("file cursor was not rewound, got %q", string(rewound))
	}
}

func TestOptionalFormFileErrorAllowsMissingFile(t *testing.T) {
	if err := OptionalFormFileError(http.ErrMissingFile); err != nil {
		t.Fatalf("OptionalFormFileError() error = %v, want nil", err)
	}
}

func TestOptionalFormFileErrorRejectsMalformedMultipart(t *testing.T) {
	if err := OptionalFormFileError(http.ErrMissingBoundary); !errors.Is(err, ErrUploadParse) {
		t.Fatalf("OptionalFormFileError() error = %v, want ErrUploadParse", err)
	}
}

func TestRequiredFormFileErrorRejectsMissingFile(t *testing.T) {
	if err := RequiredFormFileError(http.ErrMissingFile); !errors.Is(err, ErrMissingPDF) {
		t.Fatalf("RequiredFormFileError() error = %v, want ErrMissingPDF", err)
	}
}

func errorsIs(err, target error) bool {
	return errors.Is(err, target)
}
