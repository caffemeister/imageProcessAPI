package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// POST
	r.Post("/upload", app.handleFileUpload)
	r.Post("/upscale/{fileID}", app.handleUpscale)

	// DELETE
	r.Delete("/files/{fileID}", app.handleDeleteFileByID)

	// GET
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(usageInfo)) })
	r.Get("/files", app.handleGetAllFiles)
	r.Get("/files/{fileID}", app.handleGetFileByID)
	return r
}

func (app *Config) handleUpscale(w http.ResponseWriter, r *http.Request) {
	// retrieve fileid from request
	fileID, err := strconv.Atoi(chi.URLParam(r, "fileID"))
	if err != nil {
		log.Println(err)
		app.respondJSON(w, http.StatusInternalServerError, "Error retrieving fileID parameter!", "")
		return
	}

	// 3 sec timeout for db query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// query the db for file
	query := "SELECT filename FROM uploads WHERE id = $1"
	row := app.Connection.QueryRow(ctx, query, fileID)
	var filename string
	err = row.Scan(&filename)
	if err != nil {
		log.Println(err)
		app.respondJSON(w, http.StatusInternalServerError, "failed to retrieve filename from db", "")
		return
	}

	// create a JSON structure to send to python
	jsonMap := map[string]string{
		"filename": filename,
	}

	// add data to structure
	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		log.Println("error marshalling data:", err)
		app.respondJSON(w, http.StatusInternalServerError, "failed to prepare data as JSON", "")
		return
	}

	// create a new request to send to python
	req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("error creating request:", err)
		app.respondJSON(w, http.StatusInternalServerError, "failed to create request", "")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println("error contacting python service:", err)
		app.respondJSON(w, http.StatusInternalServerError, "failed to contact python service", "")
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		app.respondJSON(w, http.StatusInternalServerError, "bad status received from python service", filename)
	}

	app.respondJSON(w, http.StatusOK, "file upscaled successfully!", filename)
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

	// 3 second time limit for db query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// put the filename into DB
	query := "INSERT INTO uploads (filename) VALUES ($1)"
	_, err = app.Connection.Exec(ctx, query, sanitizedFilename)
	if err != nil {
		log.Println(err)
		app.respondJSON(w, http.StatusInternalServerError, "Error saving file to database", sanitizedFilename)
		return
	}

	// report status to user
	app.respondJSON(w, http.StatusOK, "File uploaded successfully", sanitizedFilename)
}

// handles GET to "/files"
func (app *Config) handleGetAllFiles(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, filename FROM uploads WHERE 1=1"

	// 3 second time limit for db query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := app.Connection.Query(ctx, query)
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "Failed to retrieve files from database", "")
	}
	defer rows.Close()

	var lines []string

	for rows.Next() {
		var filename string
		var fileID int
		err := rows.Scan(&fileID, &filename)
		if err != nil {
			app.respondJSON(w, http.StatusInternalServerError, "Error reading row data", "")
			return
		}
		line := filename + " " + "[" + strconv.Itoa(fileID) + "]"
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

	// 3 second time limit for db query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT filename FROM uploads WHERE id = $1"
	row := app.Connection.QueryRow(ctx, query, fileID)

	var filename string
	err = row.Scan(&filename)
	if err != nil {
		app.respondJSON(w, http.StatusNotFound, "failed to locate file with this ID", "")
		return
	}

	app.respondJSON(w, http.StatusOK, "file found", filename)
}

func (app *Config) handleDeleteFileByID(w http.ResponseWriter, r *http.Request) {
	// get the fileID to delete
	fileID, err := strconv.Atoi(chi.URLParam(r, "fileID"))
	if err != nil {
		app.respondJSON(w, http.StatusBadRequest, "Invalid file ID", "")
		return
	}

	// find the file in DB
	var filename string

	// 3 second time limit for db query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT filename FROM uploads WHERE id = $1"
	err = app.Connection.QueryRow(ctx, query, fileID).Scan(&filename)
	if err != nil {
		if err == sql.ErrNoRows {
			app.respondJSON(w, http.StatusNotFound, "File not found", "")
		} else {
			app.respondJSON(w, http.StatusInternalServerError, "Error querying database", "")
		}
		return
	}

	// remove locally
	filePath := "./../uploads/" + filename
	err = os.Remove(filePath)
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "Failed to remove file from filesystem", "")
		return
	}

	// another 3 second time limit for db query (2nd time so it
	// wouldn't share the 3 sec from the initial 3 sec context)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	// remove in DB
	deleteQuery := "DELETE FROM uploads WHERE id = $1"
	_, err = app.Connection.Exec(ctx2, deleteQuery, fileID)
	if err != nil {
		app.respondJSON(w, http.StatusInternalServerError, "Failed to delete file record from database", "")
		return
	}

	app.respondJSON(w, http.StatusOK, "File successfully deleted", "")
}
