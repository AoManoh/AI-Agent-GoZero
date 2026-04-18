package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"GoZero-AI/mcp/types/mcp"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func main() {
	var (
		pdfPath  string
		endpoint string
		output   string
	)

	flag.StringVar(&pdfPath, "pdf", filepath.Join("notes", "国务院令第493号：生产安全事故报告和调查处理条例 - 国家煤矿安全监察局.pdf"), "Path to the PDF file to parse")
	flag.StringVar(&endpoint, "endpoint", "127.0.0.1:8080", "MCP gRPC endpoint")
	flag.StringVar(&output, "out", "", "Optional output file path")
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
	content, err := extractText(client, pdfFile, filepath.Base(pdfPath))
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

func extractText(client mcp.PdfProcessorClient, file multipart.File, filename string) (string, error) {
	stream, err := client.ExtractTextFromPDF(context.Background())
	if err != nil {
		return "", err
	}
	defer func() {
		_ = stream.CloseSend()
	}()

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

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	if err := stream.Send(&mcp.PdfReq{
		Data: &mcp.PdfReq_Chunk{Chunk: data},
	}); err != nil {
		return "", err
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}

	if res.GetError() != "" {
		return "", fmt.Errorf("%s", res.GetError())
	}

	return res.GetContent(), nil
}
