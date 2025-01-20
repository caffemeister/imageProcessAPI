package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// handles POST to "/upload"
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// parse the r.Body and save it in memory
	err := r.ParseMultipartForm(int64(maxFileSize))
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing multipart form data: %s", err), http.StatusBadRequest)
		return
	}

	// extract the file data from memory
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("error extracting file from memory: %s", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// create the file in uploadDir
	dst, err := os.Create(uploadDir + header.Filename)
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
