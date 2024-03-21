package main

import (
	"bytes"
	"context"
	"fmt"
	examplesv1 "invoice-manager/main/proto/examples/v1"
	"invoice-manager/main/proto/examples/v1/examplesv1connect"
	file_uploadv1 "invoice-manager/main/proto/file_upload/v1"
	file_uploadv1connect "invoice-manager/main/proto/file_upload/v1/file_uploadv1connect"
	pingv1 "invoice-manager/main/proto/ping/v1"
	pingv1connect "invoice-manager/main/proto/ping/v1/pingv1connect"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func handleCors(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:80",
			"http://localhost:8080",
		},
		AllowedMethods: []string{
			"GET",  // for Connect
			"POST", // for all protocols,
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",             // for all protocols
			"Connect-Protocol-Version", // for Connect
			"Connect-Timeout-Ms",       // for Connect
			"Grpc-Timeout",             // for gRPC-web
			"X-Grpc-Web",               // for gRPC-web
			"X-User-Agent",             // for all protocols
		},
		ExposedHeaders: []string{
			"Grpc-Status",             // for gRPC-web
			"Grpc-Message",            // for gRPC-web
			"Grpc-Status-Details-Bin", // for gRPC-web
		},
		MaxAge: 7200, // 2 hours in seconds
	}).Handler(h)
}

type PingServer struct {
	pingv1connect.UnimplementedPingServiceHandler
}

type ExamplesServer struct {
	examplesv1connect.UnimplementedExampleServiceHandler
}

type FileUploadServer struct {
	file_uploadv1connect.UnimplementedFileUploadServiceHandler
}

type File struct {
	FilePath   string
	buffer     *bytes.Buffer
	OutputFile *os.File
}

func (f *File) SetFile(fileName string, path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	f.FilePath = filepath.Join(path, fileName)
	file, err := os.Create(f.FilePath)
	if err != nil {
		return err
	}

	f.OutputFile = file
	return nil
}

func (fus *FileUploadServer) UploadFile(
	ctx context.Context,
	stream *connect.ClientStream[file_uploadv1.FileUploadRequest],
) (*connect.Response[file_uploadv1.FileUploadResponse], error) {
	log.Println("Hit FileUploadServer::UploadFile()")
	file := &File{buffer: &bytes.Buffer{}}
	fileSize := uint32(0)

	for stream.Receive() {
		if file.FilePath == "" {
			file.SetFile(stream.Msg().GetFileName(), "client_files")
		}

		chunk := stream.Msg().GetChunk()
		fileSize = uint32(len(chunk))
		log.Println("Received a chunk with size: %d", fileSize)
		if _, err := file.OutputFile.Write(chunk); err != nil {
			log.Println("An error occured while writing a chunk", err)
			return nil, err
		}
	}

	if err := stream.Err(); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	fileName := filepath.Base(file.FilePath)
	log.Println("saved file: %s, size %d", fileName, fileSize)
	res := connect.NewResponse(&file_uploadv1.FileUploadResponse{
		FileName: fileName,
		Size:     fileSize,
	})
	return res, nil

}

func (ps *PingServer) Ping(
	ctx context.Context,
	req *connect.Request[pingv1.PingRequest],
) (*connect.Response[pingv1.PingResponse], error) {
	log.Println("Hit PingServer::Ping()")
	res := connect.NewResponse(&pingv1.PingResponse{
		Text:   req.Msg.Text,
		Number: req.Msg.Number,
	})
	res.Header().Set("Some-Other-Header", "hello!")
	return res, nil
}

func (es *ExamplesServer) Index(
	ctx context.Context,
	req *connect.Request[examplesv1.IndexRequest],
) (*connect.Response[examplesv1.Examples], error) {
	log.Println("Hit ExamplesServer::Index()")
	examples := examplesv1.Examples{Examples: []*examplesv1.Example{}}
	res := connect.NewResponse(&examples)
	return res, nil
}

func FileUploadHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(10 << 20) // 10mb
	file, handler, err := req.FormFile("file")
	if err != nil {
		log.Fatal("got error", err)
	}

	log.Println("got file", file)
	defer file.Close()

	log.Println("1")
	path := "uploaded_files"
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println("err")
		log.Fatal(err)
	}
	file_path := filepath.Join(path, handler.Filename)
	dst, err := os.Create(file_path)
	if err != nil {
		log.Println("error creating file", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("uploaded file")
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(pingv1connect.NewPingServiceHandler(&PingServer{}))
	mux.Handle(examplesv1connect.NewExampleServiceHandler(&ExamplesServer{}))
	mux.Handle(file_uploadv1connect.NewFileUploadServiceHandler(&FileUploadServer{}))
	mux.HandleFunc("/file-upload", FileUploadHandler)
	handler := h2c.NewHandler(mux, &http2.Server{})
	handler = handleCors(handler)

	httpAddress := "localhost:9002"
	fmt.Println("HTTP server listening on", httpAddress)
	err := http.ListenAndServe(httpAddress, handler)
	if err != nil {
		log.Fatal("Failed to start a HTTP server:", err)
	}
}
