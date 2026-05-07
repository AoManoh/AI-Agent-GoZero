package logic

import (
	"context"
	"testing"

	"GoZero-AI/internal/mcpsecurity"
	"GoZero-AI/mcp/internal/config"
	"GoZero-AI/mcp/internal/svc"
	"GoZero-AI/mcp/types/mcp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakePDFServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakePDFServerStream) Context() context.Context {
	return s.ctx
}

func (s *fakePDFServerStream) Recv() (*mcp.PdfReq, error) {
	return nil, nil
}

func (s *fakePDFServerStream) SendAndClose(*mcp.PdfRes) error {
	return nil
}

func TestAuthorizeStreamAllowsEmptyToken(t *testing.T) {
	logic := NewExtractTextFromPDFLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{},
	})

	err := logic.authorizeStream(&fakePDFServerStream{ctx: context.Background()})
	if err != nil {
		t.Fatalf("authorizeStream() error = %v, want nil", err)
	}
}

func TestAuthorizeStreamRejectsMissingToken(t *testing.T) {
	logic := NewExtractTextFromPDFLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{PDF: config.PDFConfig{AuthToken: "secret-token"}},
	})

	err := logic.authorizeStream(&fakePDFServerStream{ctx: context.Background()})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("authorizeStream() code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestAuthorizeStreamAcceptsIncomingToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		mcpsecurity.AuthTokenMetadataKey,
		"secret-token",
	))
	logic := NewExtractTextFromPDFLogic(ctx, &svc.ServiceContext{
		Config: config.Config{PDF: config.PDFConfig{AuthToken: "secret-token"}},
	})

	err := logic.authorizeStream(&fakePDFServerStream{ctx: ctx})
	if err != nil {
		t.Fatalf("authorizeStream() error = %v, want nil", err)
	}
}

func TestCheckUploadSizeEnforcesConfiguredLimit(t *testing.T) {
	logic := NewExtractTextFromPDFLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{PDF: config.PDFConfig{MaxUploadBytes: 3}},
	})

	var received int64
	if err := logic.checkUploadSize(&received, 2); err != nil {
		t.Fatalf("checkUploadSize() first chunk error = %v, want nil", err)
	}
	err := logic.checkUploadSize(&received, 2)
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("checkUploadSize() code = %v, want ResourceExhausted", status.Code(err))
	}
	if received != 4 {
		t.Fatalf("received bytes = %d, want 4", received)
	}
}
