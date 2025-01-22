package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var lc = "http://localhost:8080"

func TestHandleFileUpload(t *testing.T) {
	pathToTestImage := "./../testimage.png"
	app := &Config{
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
		Uploads:           []fileUpload{},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/upload", app.handleFileUpload)

	fileToUpload, err := os.Open(pathToTestImage)
	if err != nil {
		t.Error(err)
	}
	defer fileToUpload.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", "testimage.png")
	if err != nil {
		t.Error(err)
	}

	_, err = io.Copy(part, fileToUpload)
	if err != nil {
		t.Fatal("Failed to copy file contents:", err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, lc+"/upload", &requestBody)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("GET to '/upload' -> expected code 200, got code", rr.Code)
	}
}
