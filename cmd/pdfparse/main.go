package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"GoZero-AI/internal/pdfgrpc"
	"GoZero-AI/mcp/types/mcp"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func main() {
	var (
		pdfPath   string
		endpoint  string
		output    string
		authToken string
		maxBytes  int64
	)

	flag.StringVar(&pdfPath, "pdf", filepath.Join("notes", "国务院令第493号：生产安全事故报告和调查处理条例 - 国家煤矿安全监察局.pdf"), "Path to the PDF file to parse")
	flag.StringVar(&endpoint, "endpoint", "127.0.0.1:8080", "MCP gRPC endpoint")
	flag.StringVar(&output, "out", "", "Optional output file path")
	flag.StringVar(&authToken, "token", "", "MCP service auth token")
	flag.Int64Var(&maxBytes, "max-bytes", 50*1024*1024, "Maximum PDF upload bytes")
	flag.Parse()

	if pdfPath == "" {
		log.Fatal("pdf path is required")
	}

	file, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("failed to open pdf: %v", err)
	}
	defer file.Close()

	client := newPdfClient(endpoint)
	var pdfFile multipart.File = file
	content, err := extractText(context.Background(), client, pdfFile, filepath.Base(pdfPath), authToken, maxBytes)
	if err != nil {
		log.Fatalf("extract text failed: %v", err)
	}

	outPath := output
	if outPath == "" {
		base := pdfPath[:len(pdfPath)-len(filepath.Ext(pdfPath))]
		outPath = base + ".txt"
	}

	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		log.Fatalf("write output failed: %v", err)
	}

	fmt.Printf("PDF content written to %s\n", outPath)
}

func newPdfClient(endpoint string) mcp.PdfProcessorClient {
	cli := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{endpoint},
		NonBlock:  true,
	}, zrpc.WithDialOption(grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(50*1024*1024),
		grpc.MaxCallRecvMsgSize(50*1024*1024),
	)))

	return mcp.NewPdfProcessorClient(cli.Conn())
}

func extractText(ctx context.Context, client mcp.PdfProcessorClient, file multipart.File, filename, authToken string, maxBytes int64) (string, error) {
	return pdfgrpc.ExtractTextWithOptions(ctx, func(ctx context.Context) (pdfgrpc.ClientStream, error) {
		return client.ExtractTextFromPDF(ctx)
	}, file, filename, pdfgrpc.ExtractOptions{
		AuthToken:      authToken,
		MaxUploadBytes: maxBytes,
	})
}
