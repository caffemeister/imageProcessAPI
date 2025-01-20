package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var uploadDir = "./../uploads/"

func main() {
	checkUploadDir()

	// create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		// parse the r.Body and save it in memory
		err := r.ParseMultipartForm(10 << 20) // ~10 mb
		if err != nil {
			http.Error(w, "error parsing multipart form data", http.StatusBadRequest)
			return
		}

		// extract the file data from memory
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Println("error extracting file from memory", err)
			http.Error(w, "error extracting file from memory", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// create the file in uploadDir
		dst, err := os.Create(uploadDir + header.Filename)
		if err != nil {
			log.Println("error creating file", err)
			http.Error(w, "Error creating file", http.StatusInternalServerError)
			return
		}

		// copy file contents to file
		_, err = io.Copy(dst, file)
		if err != nil {
			log.Println("error saving file", err)
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// report status to user
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File uploaded successfully"))
	})

	http.ListenAndServe(":8080", r)
}

// Checks if upload dir exists, if not, creates it
func checkUploadDir() {
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
