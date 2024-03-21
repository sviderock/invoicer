package main

import (
	"fmt"
	"log"
	"net/http"

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

func main() {
	mux := http.NewServeMux()

	// gRPC services
	mux.Handle(PingServiceHandler())

	// Rest API
	mux.Handle("/"+STATIC_PATH+"/", http.FileServer(http.Dir('.')))
	mux.HandleFunc("/file-upload", UploadFile)
	mux.HandleFunc("/get-files", ReadUploadedFiles)
	mux.HandleFunc("/update-file", UpdateFile)

	handler := h2c.NewHandler(mux, &http2.Server{})
	handler = handleCors(handler)

	httpAddress := "localhost:9002"
	fmt.Println("HTTP server listening on", httpAddress)
	err := http.ListenAndServe(httpAddress, handler)
	if err != nil {
		log.Fatal("Failed to start a HTTP server:", err)
	}
}
