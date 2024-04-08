package main

import (
	"fmt"
	"invoice-manager/main/internal/constants"
	"invoice-manager/main/internal/ping"
	"invoice-manager/main/internal/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func handleCors() *cors.Cors {
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
			"PATCH",
			"PUT",
			"DELETE",
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
	})
}

type Api struct {
	TemplatesApi *template.TemplateApi
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	api := &Api{
		TemplatesApi: template.NewTemplateApi(),
	}

	r := mux.NewRouter()

	// gRPC services
	r.Handle(ping.PingServiceHandler())

	// Rest API
	r.PathPrefix("/" + template.STATIC_DIR).Handler(http.FileServer(http.Dir(".")))
	r.HandleFunc("/templates", api.TemplatesApi.GetTemplatesList).Methods("GET")
	r.HandleFunc("/templates", api.TemplatesApi.UploadFile).Methods("POST")
	r.HandleFunc("/templates/{id:[0-9]+}", api.TemplatesApi.UpdateTemplate).Methods("PATCH")
	r.HandleFunc("/templates/{id:[0-9]+}", api.TemplatesApi.DeleteTemplate).Methods("DELETE")
	r.HandleFunc("/templates/{id:[0-9]+}/html", api.TemplatesApi.UpdateTemplateHtml).Methods("PUT")

	handler := h2c.NewHandler(r, &http2.Server{})
	handler = handleCors().Handler(handler)

	fmt.Println("HTTP server listening on", constants.HTTP_ADDR)
	err := http.ListenAndServe(constants.HTTP_ADDR, handler)
	if err != nil {
		log.Fatal("Failed to start a HTTP server:", err)
	}
}
