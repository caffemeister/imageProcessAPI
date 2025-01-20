package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var uploadDir = "./../uploads/"
var maxFileSize = 10 << 20 // ~10 mb
var allowedExtensions = []string{"png", "jpg", "jpeg"}

type Config struct {
	UploadDir         string
	MaxFileSize       int
	AllowedExtensions []string
}

func main() {
	app := &Config{
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}

	// check for ./uploads/
	app.checkUploadDirExists()

	// create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// POST
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	r.Post("/upload", app.handleFileUpload)

	// GET
	r.Get("/files", app.showFiles)

	http.ListenAndServe(":8080", r)
}
