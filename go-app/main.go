package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var uploadDir = "./../uploads/"
var maxFileSize = 10 << 20 // ~10 mb

func main() {
	checkUploadDirExists()

	// create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	r.Post("/upload", handleFileUpload)

	http.ListenAndServe(":8080", r)
}

// Checks if upload dir exists, if not, creates it
func checkUploadDirExists() {
	stat, err := os.Stat(uploadDir)

	if err == nil && stat.IsDir() {
		return
	} else {
		err := os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
