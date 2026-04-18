package pdfgrpc

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"GoZero-AI/mcp/types/mcp"
)

const defaultChunkSize = 64 * 1024

type ClientStream interface {
	Send(*mcp.PdfReq) error
	CloseAndRecv() (*mcp.PdfRes, error)
	CloseSend() error
}

type StreamFactory func(ctx context.Context) (ClientStream, error)

func ExtractText(ctx context.Context, openStream StreamFactory, file multipart.File, filename string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
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
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		n, readErr := file.Read(buffer)
		if n > 0 {
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
