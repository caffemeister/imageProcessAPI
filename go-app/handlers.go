package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type fileUpload struct {
	Name string
	ID   int
}

// handles POST to "/upload"
func (app *Config) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// parse the r.Body and save it in memory
	err := r.ParseMultipartForm(int64(app.MaxFileSize))
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing multipart form data: %s", err), http.StatusBadRequest)
		return
	}

	// extract the file data from memory
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("error extracting file from memory: %s", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// protect against problematic filenames
	sanitizedFilename := strings.ReplaceAll(header.Filename, "../", "")
	header.Filename = sanitizedFilename

	// check if uploaded file is an allowed image type
	if !app.isValidImageExtension(app.getFileExtension(header.Filename)) {
		http.Error(w, "File type is not allowed!", http.StatusBadRequest)
		return
	}

	// check if uploaded file size isn't greater than allowed size
	if header.Size > int64(app.MaxFileSize) {
		http.Error(w, "File is too large!", http.StatusBadRequest)
		return
	}

	// create the file in uploadDir
	dst, err := os.Create(app.UploadDir + header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating file: %s", err), http.StatusInternalServerError)
		return
	}

	// copy file contents to file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving file: %s", err), http.StatusInternalServerError)
		return
	}

	// report status to user
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

// handles GET to "/files"
func (app *Config) showFiles(w http.ResponseWriter, r *http.Request) {
	var filenames []string

	files, err := os.ReadDir(app.UploadDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed reading local files: %s", err), http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(filenames, ", ")))
}
