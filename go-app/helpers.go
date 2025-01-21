package main

import (
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
		app.assignIDs()
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
	return len(app.Uploads)
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

// Assigns IDs to files inside app.uploadDir
func (app *Config) assignIDs() {
	f, err := os.Open(app.UploadDir)
	if err != nil {
		log.Printf("failed to get IDs -> os.Open, error: %s", err)
		return
	}
	defer f.Close()

	files, err := f.ReadDir(0)
	if err != nil {
		log.Printf("failed to get IDs -> f.ReadDir, error: %s", err)
		return
	}

	app.Uploads = nil
	for id, file := range files {
		app.Uploads = append(app.Uploads, fileUpload{
			Filename: file.Name(),
			ID:       id,
		})
	}
}

func (app *Config) listenForFileChanges(doneChan chan bool) error {
	initialStat, err := os.Stat(app.UploadDir)
	if err != nil {
		return err
	}

	for {
		select {
		case <-doneChan:
			log.Println("stopping listenForFileChanges...")
			return nil
		default:
			stat, err := os.Stat(app.UploadDir)
			if err != nil {
				return err
			}

			if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
				app.assignIDs()
				initialStat = stat
				log.Println("noticed file changes, ran assignIDs()...")
			}

			time.Sleep(1 * time.Second)
		}
	}
}
