package pdfgrpc

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"GoZero-AI/internal/mcpsecurity"
	"GoZero-AI/mcp/types/mcp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const defaultChunkSize = 64 * 1024

var ErrUploadTooLarge = errors.New("PDF 文件超过大小限制")

type ExtractOptions struct {
	AuthToken      string
	MaxUploadBytes int64
}

type ClientStream interface {
	Send(*mcp.PdfReq) error
	CloseAndRecv() (*mcp.PdfRes, error)
	CloseSend() error
}

type StreamFactory func(ctx context.Context) (ClientStream, error)

func IsUploadTooLarge(err error) bool {
	return errors.Is(err, ErrUploadTooLarge) || status.Code(err) == codes.ResourceExhausted
}

func ExtractText(ctx context.Context, openStream StreamFactory, file multipart.File, filename string) (string, error) {
	return ExtractTextWithOptions(ctx, openStream, file, filename, ExtractOptions{})
}

func ExtractTextWithOptions(ctx context.Context, openStream StreamFactory, file multipart.File, filename string, options ExtractOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if options.AuthToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, mcpsecurity.AuthTokenMetadataKey, options.AuthToken)
	}

	stream, err := openStream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.CloseSend()

	if err := stream.Send(&mcp.PdfReq{
		Data: &mcp.PdfReq_Metadata{
			Metadata: &mcp.Metadata{
				Filename: filename,
				MimeType: "application/pdf",
			},
		},
	}); err != nil {
		return "", err
	}

	buffer := make([]byte, defaultChunkSize)
	var totalBytes int64
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		n, readErr := file.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
			if options.MaxUploadBytes > 0 && totalBytes > options.MaxUploadBytes {
				return "", ErrUploadTooLarge
			}

			chunk := make([]byte, n)
			copy(chunk, buffer[:n])
			if err := stream.Send(&mcp.PdfReq{
				Data: &mcp.PdfReq_Chunk{
					Chunk: chunk,
				},
			}); err != nil {
				return "", err
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", readErr
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}
	if res.Error != "" {
		return "", errors.New(res.Error)
	}

	return res.Content, nil
}
