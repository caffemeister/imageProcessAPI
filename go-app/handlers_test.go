package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const lc = "http://localhost:8080"

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
		t.Fatal(err)
	}
	defer fileToUpload.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", "testimage.png")
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, fileToUpload)
	if err != nil {
		t.Fatal("Failed to copy file contents:", err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, lc+"/upload", &requestBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("POST to '/upload' -> expected code 200, got code", rr.Code)
	}

	expected := "File uploaded successfully"
	if !bytes.Contains(rr.Body.Bytes(), []byte(expected)) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleGetAllFiles(t *testing.T) {
	app := &Config{
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
		Uploads:           []fileUpload{},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/files", app.handleGetAllFiles)

	app.assignIDs()

	req, err := http.NewRequest(http.MethodGet, lc+"/files", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var response jsonResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Failed to unmarshal response body:", err)
	}

	if response.Status != http.StatusOK {
		t.Error("GET to '/files' -> expected code 200, got code", rr.Code)
	}

	expected := "testimage.png"
	if !strings.Contains(response.Message, expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleGetFileByID(t *testing.T) {
	app := &Config{
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
		Uploads:           []fileUpload{{ID: 1, Filename: "testimage.png"}},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/files/{fileID}", app.handleGetFileByID)
	app.assignIDs()

	req, err := http.NewRequest(http.MethodGet, lc+"/files/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("GET to '/files/1' -> expected code 200, got code", rr.Code)
	}

	expected := "file found"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}
}
