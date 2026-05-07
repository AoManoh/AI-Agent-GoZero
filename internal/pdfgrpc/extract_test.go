package pdfgrpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"GoZero-AI/internal/mcpsecurity"
	"GoZero-AI/mcp/types/mcp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testMultipartFile struct {
	*bytes.Reader
}

func (f *testMultipartFile) Close() error {
	return nil
}

type fakeStream struct {
	sent      []*mcp.PdfReq
	closeResp *mcp.PdfRes
	sendErr   error
	recvErr   error
	closed    bool
}

func (s *fakeStream) Send(req *mcp.PdfReq) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.sent = append(s.sent, req)
	return nil
}

func (s *fakeStream) CloseAndRecv() (*mcp.PdfRes, error) {
	if s.recvErr != nil {
		return nil, s.recvErr
	}
	if s.closeResp == nil {
		return &mcp.PdfRes{}, nil
	}
	return s.closeResp, nil
}

func (s *fakeStream) CloseSend() error {
	s.closed = true
	return nil
}

func TestExtractTextStreamsMetadataAndChunks(t *testing.T) {
	payload := bytes.Repeat([]byte("a"), defaultChunkSize*2+123)
	file := &testMultipartFile{Reader: bytes.NewReader(payload)}
	stream := &fakeStream{closeResp: &mcp.PdfRes{Content: "ok"}}

	content, err := ExtractText(context.Background(), func(ctx context.Context) (ClientStream, error) {
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		return stream, nil
	}, file, "resume.pdf")
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if content != "ok" {
		t.Fatalf("ExtractText() content = %q, want ok", content)
	}
	if !stream.closed {
		t.Fatal("expected CloseSend to be called")
	}
	if len(stream.sent) != 4 {
		t.Fatalf("sent request count = %d, want 4", len(stream.sent))
	}

	metadata := stream.sent[0].GetMetadata()
	if metadata == nil || metadata.Filename != "resume.pdf" || metadata.MimeType != "application/pdf" {
		t.Fatalf("unexpected metadata payload: %#v", metadata)
	}

	chunk1 := stream.sent[1].GetChunk()
	chunk2 := stream.sent[2].GetChunk()
	chunk3 := stream.sent[3].GetChunk()
	if len(chunk1) != defaultChunkSize || len(chunk2) != defaultChunkSize || len(chunk3) != 123 {
		t.Fatalf("unexpected chunk sizes: %d, %d, %d", len(chunk1), len(chunk2), len(chunk3))
	}
}

func TestExtractTextAttachesAuthToken(t *testing.T) {
	file := &testMultipartFile{Reader: bytes.NewReader([]byte("payload"))}
	stream := &fakeStream{closeResp: &mcp.PdfRes{Content: "ok"}}

	_, err := ExtractTextWithOptions(context.Background(), func(ctx context.Context) (ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		values := md.Get(mcpsecurity.AuthTokenMetadataKey)
		if len(values) != 1 || values[0] != "secret-token" {
			t.Fatalf("auth token metadata = %#v, want secret-token", values)
		}
		return stream, nil
	}, file, "resume.pdf", ExtractOptions{AuthToken: "secret-token"})
	if err != nil {
		t.Fatalf("ExtractTextWithOptions() error = %v", err)
	}
}

func TestExtractTextStopsAtMaxUploadBytes(t *testing.T) {
	file := &testMultipartFile{Reader: bytes.NewReader([]byte("payload"))}
	stream := &fakeStream{closeResp: &mcp.PdfRes{Content: "ok"}}

	_, err := ExtractTextWithOptions(context.Background(), func(ctx context.Context) (ClientStream, error) {
		return stream, nil
	}, file, "resume.pdf", ExtractOptions{MaxUploadBytes: 3})
	if !errors.Is(err, ErrUploadTooLarge) {
		t.Fatalf("ExtractTextWithOptions() error = %v, want ErrUploadTooLarge", err)
	}
	if !stream.closed {
		t.Fatal("expected CloseSend to be called after size limit failure")
	}
	if len(stream.sent) != 1 {
		t.Fatalf("sent request count = %d, want metadata only", len(stream.sent))
	}
}

func TestIsUploadTooLargeMatchesGrpcResourceExhausted(t *testing.T) {
	err := status.Error(codes.ResourceExhausted, "PDF 文件超过大小限制")
	if !IsUploadTooLarge(err) {
		t.Fatal("expected ResourceExhausted gRPC error to be upload-too-large")
	}
}

func TestExtractTextReturnsContextErrorBeforeOpen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	file := &testMultipartFile{Reader: bytes.NewReader([]byte("%PDF-1.4"))}
	opened := false
	_, err := ExtractText(ctx, func(ctx context.Context) (ClientStream, error) {
		opened = true
		return &fakeStream{}, nil
	}, file, "resume.pdf")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExtractText() error = %v, want context.Canceled", err)
	}
	if opened {
		t.Fatal("expected stream not to open after context cancellation")
	}
}

func TestExtractTextReturnsServerError(t *testing.T) {
	file := &testMultipartFile{Reader: bytes.NewReader([]byte("payload"))}
	stream := &fakeStream{closeResp: &mcp.PdfRes{Error: "extract failed"}}

	_, err := ExtractText(context.Background(), func(ctx context.Context) (ClientStream, error) {
		return stream, nil
	}, file, "resume.pdf")
	if err == nil || err.Error() != "extract failed" {
		t.Fatalf("ExtractText() error = %v, want extract failed", err)
	}
}

func TestExtractTextPropagatesReadError(t *testing.T) {
	file := &errorMultipartFile{}
	stream := &fakeStream{}

	_, err := ExtractText(context.Background(), func(ctx context.Context) (ClientStream, error) {
		return stream, nil
	}, file, "resume.pdf")
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("ExtractText() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

type errorMultipartFile struct{}

func (f *errorMultipartFile) Read(p []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (f *errorMultipartFile) ReadAt(p []byte, off int64) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
func (f *errorMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
func (f *errorMultipartFile) Close() error { return nil }
