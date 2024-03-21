package main

import (
	"encoding/json"
	"errors"
	"fmt"
	pb "invoice-manager/main/proto"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/bimg"
)

const (
	STATIC_PATH    = "static"
	UPLOAD_DIR     = "uploaded"
	THUMBNAIL_PATH = "thumbnail.jpg"
)

var (
	FILE_UPLOAD_PATH = filepath.Join(STATIC_PATH, UPLOAD_DIR)
)

func GetFile(dir fs.DirEntry) *pb.File {
	var file pb.File
	file_id := dir.Name()
	file_dir := filepath.Join(FILE_UPLOAD_PATH, file_id)
	files, err := os.ReadDir(file_dir)

	if err != nil {
		log.Println("Failed to read directory ", file_dir)
		return nil
	}

	for _, dir_file := range files {
		file_info, err := dir_file.Info()
		if err != nil {
			log.Println("Failed to read file", dir_file.Name())
			return nil
		}

		if file_info.Name() == THUMBNAIL_PATH {
			file.Thumbnail = "/" + path.Join(FILE_UPLOAD_PATH, file_id, THUMBNAIL_PATH)
			continue
		}

		file.Id = file_id
		file.Name = file_info.Name()
		file.Ext = filepath.Ext(file_info.Name())
		file.Size = uint32(file_info.Size())
		file.Path = "/" + path.Join(FILE_UPLOAD_PATH, file_id, file_info.Name())
	}

	return &file
}

func UploadFile(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(10 << 20) // 10mb
	file, handler, err := req.FormFile("file")
	log.Println(handler.Filename)
	if err != nil {
		log.Println("Couldn't read from req.FormFile(\"file\")")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	file_id := fmt.Sprint(time.Now().UnixNano())
	file_dir := filepath.Join(FILE_UPLOAD_PATH, file_id)
	err = os.MkdirAll(file_dir, os.ModePerm)
	if err != nil {
		log.Printf("Error creating %s directory", file_dir)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file_ext := filepath.Ext(handler.Filename)
	file_path := filepath.Join(file_dir, handler.Filename)
	file_thumbnail_path := filepath.Join(file_dir, THUMBNAIL_PATH)

	if file_ext != ".pdf" {
		log.Println("File is not a PDF")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dst, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Println("Error creating file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Println("Error copying file into destination", dst)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	thumbnail_buffer, err := bimg.Read(file_path)
	if err != nil {
		log.Println("Error reading buffer for thumbnail", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	thumbnail, err := bimg.NewImage(thumbnail_buffer).Convert(bimg.JPEG)
	if err != nil {
		log.Println("Error converting thumbnail", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bimg.Write(file_thumbnail_path, thumbnail)
	if err != nil {
		log.Println("Error creating thumbnail", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploaded_file := pb.File{
		Id:        file_id,
		Name:      handler.Filename,
		Ext:       file_ext,
		Path:      "/" + file_path,
		Size:      uint32(handler.Size),
		Thumbnail: "/" + file_thumbnail_path,
	}

	json, err := json.Marshal(pb.FileUploadResponse{File: &uploaded_file})

	if err != nil {
		log.Fatalln("Error while encoding json", err)
	}

	log.Println("uploaded file", file_path)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func ReadUploadedFiles(w http.ResponseWriter, req *http.Request) {
	resp := pb.GetFilesResponse{Files: []*pb.File{}}
	dirs, err := os.ReadDir(FILE_UPLOAD_PATH)
	// resp.Files = append(resp.Files, file)
	for _, dir := range dirs {
		file := GetFile(dir)
		resp.Files = append(resp.Files, file)
	}

	if err != nil {
		log.Println("Error reading files", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(&resp)

	if err != nil {
		log.Println("Error encoding json", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func FindFile(id string) (*pb.File, error) {
	var found_file *pb.File
	dirs, err := os.ReadDir(FILE_UPLOAD_PATH)

	if err != nil {
		log.Print("Error reading directory", FILE_UPLOAD_PATH)
		return nil, err
	}

	for _, dir := range dirs {
		if dir.Name() == id {
			found_file = GetFile(dir)
			break
		}
	}

	if found_file == nil {
		return nil, errors.New("file not found")
	}

	return found_file, nil

}

func UpdateFile(w http.ResponseWriter, req *http.Request) {
	var body pb.UpdateFileNameRequest
	err := decodeJSONBody(w, req, body)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if body.Id == "" || body.Name == "" {
		log.Print("Some fields are empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	file_found, err := FindFile(body.Id)
	if err != nil {
		log.Println("Error occured while looking for file", err)
		return
	}

	file_path_split := strings.Split(file_found.Path, "/")
	new_file_split := make([]string, len(file_path_split))
	copy(new_file_split, file_path_split)

	new_file_name := body.Name + file_found.Ext
	new_path_split := append(new_file_split[:len(new_file_split)-1], new_file_name)
	old_path := filepath.Join(file_path_split...)
	new_path := filepath.Join(new_path_split...)

	err = os.Rename(old_path, new_path)
	if err != nil {
		log.Println("Error occured while renaming file from", old_path, "to", new_path)
		return
	}

	updated_file, err := FindFile(body.Id)
	if err != nil {
		log.Println("Error when retrieving updated file", err)
		return
	}

	jsonData, err := json.Marshal(&updated_file)
	if err != nil {
		log.Println("Error while encoding json", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func DeleteFile(w http.ResponseWriter, req *http.Request) {

}
