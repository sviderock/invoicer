package main

import (
	"encoding/json"
	"fmt"
	pb "invoice-manager/main/proto"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/bimg"
)

var (
	fileUploadPath = filepath.Join("static", "uploaded", "files")
	thumbnailsPath = filepath.Join("static", "uploaded", "thumbnails")
)

func UploadFile(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(10 << 20) // 10mb
	file, handler, err := req.FormFile("file")
	log.Println(handler.Filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	err = os.MkdirAll(fileUploadPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	file_name := fmt.Sprint(time.Now().UnixNano()) + "_" + handler.Filename
	file_name_without_ext := strings.TrimSuffix(file_name, filepath.Ext(file_name))
	file_path := filepath.Join(fileUploadPath, file_name)
	dst, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Println("error creating file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buffer, err := bimg.Read(file_path)
	if err != nil {
		log.Fatal(err)
	}
	thumbnail, err := bimg.NewImage(buffer).Convert(bimg.JPEG)
	if err != nil {
		log.Fatal(err)
	}
	thumbnail_name := file_name_without_ext + ".jpg"
	thumbnail_path := filepath.Join(thumbnailsPath, thumbnail_name)
	bimg.Write(thumbnail_path, thumbnail)

	json, err := json.Marshal(pb.FileUploadResponse{
		File: &pb.File{
			Name:      file_name,
			Path:      "/" + file_path,
			Size:      uint32(handler.Size),
			Thumbnail: "/" + thumbnail_path,
		},
	})

	if err != nil {
		log.Fatalln("Error while encoding json", err)
	}

	log.Println("uploaded file", file_path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func ReadUploadedFiles(w http.ResponseWriter, req *http.Request) {
	files, err := os.ReadDir(fileUploadPath)
	if err != nil {
		log.Fatalln(err)
	}

	resp := pb.GetFilesResponse{Files: []*pb.File{}}
	for _, file := range files {
		if file.Name() == ".gitkeep" {
			continue
		}

		file_info, err := file.Info()
		if err != nil {
			log.Fatalln(err)
		}

		resp.Files = append(resp.Files, &pb.File{
			Name: file.Name(),
			Path: "/" + fileUploadPath + "/" + file.Name(),
			Size: uint32(file_info.Size()),
		})
	}

	jsonData, err := json.Marshal(resp)

	if err != nil {
		log.Fatalln(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
