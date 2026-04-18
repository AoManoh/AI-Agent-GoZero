package svc

import (
	"context"
	"mime/multipart"

	"GoZero-AI/internal/pdfgrpc"
	"GoZero-AI/mcp/types/mcp"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type PdfClient struct {
	Client mcp.PdfProcessorClient
}

func NewPdfClient(endPoint string) *PdfClient {
	conn := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{endPoint},
		NonBlock:  true,
	}, zrpc.WithDialOption(grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(50*1024*1024),
		grpc.MaxCallRecvMsgSize(50*1024*1024),
	)))

	return &PdfClient{
		Client: mcp.NewPdfProcessorClient(conn.Conn()),
	}
}

func (c *PdfClient) ExtractTextFromPDF(ctx context.Context, file multipart.File, filename string) (string, error) {
	content, err := pdfgrpc.ExtractText(ctx, func(ctx context.Context) (pdfgrpc.ClientStream, error) {
		stream, err := c.Client.ExtractTextFromPDF(ctx)
		if err != nil {
			logx.Errorf("创建 gRPC 流式客户端失败: %v", err)
			return nil, err
		}
		return stream, nil
	}, file, filename)
	if err != nil {
		logx.Errorf("PDF 文本提取失败: %v", err)
		return "", err
	}

	return content, nil
}
