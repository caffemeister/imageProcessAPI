package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var uploadDir = "./../uploads/"
var maxFileSize = 10 << 20                                 // ~10 mb
var allowedExtensions = []string{"png", "jpg", "jpeg"}     // allowed types of files
var doneChan = make(chan bool)                             // channel for app.listenForFileChanges()
var usageInfo = "POST: /uploads\nGET: /files, /files/<id>" // for GET to "/", shows usage

type Config struct {
	UploadDir         string
	MaxFileSize       int
	AllowedExtensions []string
	Uploads           []fileUpload
}

type fileUpload struct {
	Filename string
	ID       int
}

type jsonResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	File    string `json:"file,omitempty"`
}

func main() {
	app := &Config{
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
		Uploads:           []fileUpload{},
	}

	// check for ./uploads/
	app.checkUploadDirExists()
	go app.listenForFileChanges(doneChan)
	defer close(doneChan)

	// create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// POST
	r.Post("/upload", app.handleFileUpload)
	// TODO: ADD DELETE

	// GET
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(usageInfo)) })
	r.Get("/files", app.handleGetAllFiles)
	r.Get("/files/{fileID}", app.handleGetFileByID)

	http.ListenAndServe(":8080", r)
}
