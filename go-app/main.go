package main

import (
	"net/http"
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
	r := app.routes()

	http.ListenAndServe(":8080", r)
}
