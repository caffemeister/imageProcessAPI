package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

// Checks if upload dir exists, if not, creates it
func (app *Config) checkUploadDirExists() {
	stat, err := os.Stat(app.UploadDir)

	if err == nil {
		if !stat.IsDir() {
			log.Fatalf("Path %s exists but is not a directory!", app.UploadDir)
		}
		return
	}

	if os.IsNotExist(err) {
		err := os.MkdirAll(app.UploadDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create upload directory: %s", err)
		}
	} else {
		log.Fatalf("Error checking upload directory: %s", err)
	}
}

// Returns the file extension from the filename
func (app *Config) getFileExtension(file string) string {
	extension := strings.LastIndex(file, ".")
	if extension == -1 {
		return ""
	}

	return file[extension+1:]
}

// Checks if entry is in app.AllowedExtensions
func (app *Config) isValidImageExtension(entry string) bool {
	return slices.Contains(app.AllowedExtensions, entry)
}

func (app *Config) getFileCount() int {
	var fileCount int

	// 3 second timer for query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT COUNT(*) FROM uploads WHERE 1=1"
	err := app.Connection.QueryRow(ctx, query).Scan(&fileCount)
	if err != nil {
		log.Println(err)
		return -1
	}
	return fileCount
}

func (app *Config) respondJSON(w http.ResponseWriter, status int, msg string, filename string) {
	payload := jsonResponse{
		Status:  status,
		Message: msg,
		File:    filename,
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
