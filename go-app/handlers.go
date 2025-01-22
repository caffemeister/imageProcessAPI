package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// POST
	r.Post("/upload", app.handleFileUpload)

	// DELETE
	r.Delete("/files/{fileID}", app.handleDeleteFileByID)

	// GET
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(usageInfo)) })
	r.Get("/files", app.handleGetAllFiles)
	r.Get("/files/{fileID}", app.handleGetFileByID)
	return r
}

// handles POST to "/upload"
func (app *Config) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// parse the r.Body and save it in memory
	err := r.ParseMultipartForm(int64(app.MaxFileSize))
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "error parsing multipart form data", "")
		return
	}

	// extract the file data from memory
	file, header, err := r.FormFile("file")
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "error extracting file from memory", "")
		return
	}
	defer file.Close()

	// protect against problematic filenames
	sanitizedFilename := strings.ReplaceAll(header.Filename, "../", "")
	header.Filename = sanitizedFilename

	// check if uploaded file is an allowed image type
	if !app.isValidImageExtension(app.getFileExtension(header.Filename)) {
		app.respondJSON(w, http.StatusBadRequest, "File type is not allowed!", "")
		return
	}

	// check if uploaded file size isn't greater than allowed size
	if header.Size > int64(app.MaxFileSize) {
		app.respondJSON(w, http.StatusBadRequest, "File size is too large!", "")
		return
	}

	// create the file in uploadDir
	dst, err := os.Create(app.UploadDir + header.Filename)
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "Error creating file", "")
		return
	}
	defer dst.Close()

	// copy file contents to file
	_, err = io.Copy(dst, file)
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "Error saving file data", "")
		return
	}

	app.assignIDs()

	// report status to user
	app.respondJSON(w, http.StatusOK, "File uploaded successfully", sanitizedFilename)
}

// handles GET to "/files"
func (app *Config) handleGetAllFiles(w http.ResponseWriter, r *http.Request) {
	var lines []string

	for _, upload := range app.Uploads {
		line := upload.Filename + " " + "[" + strconv.Itoa(upload.ID) + "]"
		lines = append(lines, line)
	}

	app.respondJSON(w, http.StatusOK, strings.Join(lines, ", "), "")
}

// handles GET to "/files/<fileID>"
func (app *Config) handleGetFileByID(w http.ResponseWriter, r *http.Request) {
	// retrieve fileID specified by user
	fileID, err := strconv.Atoi(chi.URLParam(r, "fileID"))
	if err != nil {
		log.Println(err)
		app.respondJSON(w, http.StatusInternalServerError, "Error retrieving fileID!", "")
		return
	}

	for id, file := range app.Uploads {
		if fileID == id {
			app.respondJSON(w, http.StatusOK, "file found", file.Filename)
			return
		}
	}
	app.respondJSON(w, http.StatusNotFound, "Error locating file by ID: file does not exist!", "")
}

func (app *Config) handleDeleteFileByID(w http.ResponseWriter, r *http.Request) {
	var fileToRemove string

	// retrieve fileID specified by user
	fileID, err := strconv.Atoi(chi.URLParam(r, "fileID"))
	if err != nil {
		log.Println(err)
		app.respondJSON(w, http.StatusInternalServerError, "Error retrieving fileID!", "")
		return
	}

	// find and remove the file, then run assignIDs
	for _, file := range app.Uploads {
		if fileID == file.ID {
			fileToRemove = file.Filename
			err = os.Remove(filepath.Join(app.UploadDir, fileToRemove))
			if err != nil {
				log.Println(err)
				app.respondJSON(w, http.StatusInternalServerError, "Failed to remove file", fileToRemove)
			}
			app.assignIDs()
			app.respondJSON(w, http.StatusOK, "File successfully deleted", fileToRemove)
			return
		}
	}

	app.respondJSON(w, http.StatusInternalServerError, "Failed to find file to delete", fileToRemove)
}
