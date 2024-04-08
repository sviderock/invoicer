package template

import (
	"encoding/json"
	"fmt"
	"invoice-manager/main/internal/helpers"
	pb "invoice-manager/main/proto"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/h2non/bimg"
	"golang.org/x/net/html"
)

type TemplateApi struct {
	templates *Templates
}

const (
	STATIC_DIR     = "static"
	THUMBNAIL_NAME = "thumbnail.jpg"
)

var (
	HTML_TEMPLATES_DIR = filepath.Join(STATIC_DIR, "templates")
	THUMBNAILS_DIR     = filepath.Join(STATIC_DIR, "thumbnails")
)

func UploadToTempFile(file multipart.File, handler *multipart.FileHeader) (*os.File, error) {
	temp_file, err := os.CreateTemp(STATIC_DIR, "tmp-uploaded-pdf-*.pdf")
	if err != nil {
		return nil, err
	}

	if filepath.Ext(handler.Filename) != ".pdf" {
		return nil, err
	}

	form_file_bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if _, err = temp_file.Write(form_file_bytes); err != nil {
		return nil, err
	}

	return temp_file, nil
}

func CreateThumbnail(temp_file *os.File) (string, error) {
	uploaded_file_path := temp_file.Name()
	thumbnail_buffer, err := bimg.Read(uploaded_file_path)
	if err != nil {
		return "", err
	}

	thumbnail, err := bimg.NewImage(thumbnail_buffer).Convert(bimg.JPEG)
	if err != nil {
		return "", err
	}

	thumbnail_id := fmt.Sprint(time.Now().UnixNano())
	thumbnail_path := filepath.Join(THUMBNAILS_DIR, thumbnail_id+"_"+THUMBNAIL_NAME)
	if err = bimg.Write(thumbnail_path, thumbnail); err != nil {
		return "", err
	}

	return thumbnail_path, nil
}

func ConvertPdfToHtml(temp_file *os.File) (template_path string, err error) {
	cwd_root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	template_name := fmt.Sprint(time.Now().UnixNano()) + ".html"
	image := "pdf2htmlex/pdf2htmlex:0.18.8.rc2-master-20200820-alpine-3.12.0-x86_64"
	cmd := fmt.Sprintf(
		"docker run -t --rm -v %s:/backend -w /backend %s --zoom 1.8 --embed CFIJO --dest-dir %s %s %s --process-outline 0 --optimize-text 1",
		cwd_root,
		image,
		HTML_TEMPLATES_DIR,
		temp_file.Name(),
		template_name,
	)

	if err = exec.Command("/bin/sh", "-c", cmd).Run(); err != nil {
		log.Println(cmd, err)
		return
	}

	template_path = path.Join(HTML_TEMPLATES_DIR, template_name)
	return
}

func (ta *TemplateApi) UploadFile(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(10 << 20) // 10mb
	form_file, handler, err := req.FormFile("file")
	if err != nil {
		fmt.Fprint(w, "Couldn't get req.FormFile(\"file\")")
	}
	defer form_file.Close()

	temp_file, err := UploadToTempFile(form_file, handler)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, "Count't create temp file")
	}
	defer os.Remove(temp_file.Name())
	defer temp_file.Close()

	thumbnail_path, err := CreateThumbnail(temp_file)
	if err != nil {
		fmt.Fprint(w, "Error creating thumbnail")
	}

	template_path, err := ConvertPdfToHtml(temp_file)
	if err != nil {
		fmt.Fprint(w, "Error converting PDF to HTML")
	}

	file_ext := filepath.Ext(handler.Filename)
	new_template, err := ta.templates.Insert(&pb.Template{
		Name:      strings.TrimSuffix(handler.Filename, file_ext),
		Ext:       file_ext,
		Size:      uint32(handler.Size),
		Path:      template_path,
		Thumbnail: thumbnail_path,
	})
	if err != nil {
		fmt.Fprint(w, "Error inserting new template into the database")
	}

	helpers.JsonResponse(w, http.StatusOK, pb.FileUploadResponse{Template: new_template.data})
}

func (ta *TemplateApi) GetTemplatesList(w http.ResponseWriter, req *http.Request) {
	templates, err := ta.templates.List()
	if err != nil {
		fmt.Fprint(w, "Error reading files")
	}

	var data []pb.Template
	for _, template := range templates {
		data = append(data, pb.Template{
			Id:        template.data.Id,
			Name:      template.data.Name,
			Ext:       template.data.Ext,
			Size:      template.data.Size,
			CreatedAt: template.data.CreatedAt,
			UpdatedAt: template.data.UpdatedAt,
			Path:      template.public_path,
			Thumbnail: template.public_thumbnail_path,
		})
	}

	helpers.JsonResponse(w, http.StatusOK, data)
}

func GetId(w http.ResponseWriter, req *http.Request) int {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Fprint(w, "Invalid template ID")
	}
	return id
}

func (ta *TemplateApi) UpdateTemplate(w http.ResponseWriter, req *http.Request) {
	var body pb.UpdateTemplateRequest
	err := json.NewDecoder(req.Body).Decode(&body)
	id := GetId(w, req)

	if body.GetName() == "" {
		fmt.Fprintf(w, "Name field is empty")
	}

	updated_template, err := ta.templates.UpdateName(id, body.GetName())
	if err != nil {
		fmt.Fprint(w, "Template couldn't be updated")
	}

	helpers.JsonResponse(w, http.StatusOK, &updated_template)
}

func (ta *TemplateApi) UpdateTemplateHtml(w http.ResponseWriter, req *http.Request) {
	req.Header.Set("Accept", "text/html; charset=utf-8")
	b, err := io.ReadAll(req.Body)
	html_string := string(b)
	if err != nil {
		fmt.Fprint(w, "Error reading HTML from request")
	}

	if _, err := html.Parse(strings.NewReader(html_string)); err != nil {
		fmt.Fprint(w, "HTML seems to be invalid")
	}

	id := GetId(w, req)
	template, err := ta.templates.Retrieve(id)
	if err != nil {
		fmt.Fprint(w, "Coudln't retrieve template")
	}

	if err := os.WriteFile(template.data.Path, []byte(html_string), 0660); err != nil {
		fmt.Fprint(w, "Coudln't write into HTML file")
	}
}

func (ta *TemplateApi) DeleteTemplate(w http.ResponseWriter, req *http.Request) {
	id := GetId(w, req)
	if err := ta.templates.Delete(id); err != nil {
		fmt.Fprint(w, "Error while deleting the template")
	}
}

func NewTemplateApi() *TemplateApi {
	if err := os.MkdirAll(HTML_TEMPLATES_DIR, os.ModePerm); err != nil {
		fmt.Print("Error creating HTML templates directory")
	}

	if err := os.MkdirAll(THUMBNAILS_DIR, os.ModePerm); err != nil {
		fmt.Print("Error creating thumbnails directory")
	}

	ts, err := NewTemplates()
	if err != nil {
		fmt.Print(err)
	}

	return &TemplateApi{templates: ts}
}
