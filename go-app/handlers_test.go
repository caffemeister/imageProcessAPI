package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5"
)

const lc = "http://localhost:8080"

func TestHandleFileUpload(t *testing.T) {
	pathToTestImage := "./../testimage.png"
	app := &Config{
		Connection:        connectToDBTest(), // separate connection to db for tests
		UploadDir:         "../uploads/",     // since not in docker, dir is different
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}
	r := app.routes()
	defer app.Connection.Close(context.Background())

	// open the file
	fileToUpload, err := os.Open(pathToTestImage)
	if err != nil {
		t.Fatal(err)
	}
	defer fileToUpload.Close()

	// create an empty requestBody
	var requestBody bytes.Buffer
	// and populate it with an empty multipart/data-file form
	writer := multipart.NewWriter(&requestBody)

	// create a multipart/data-file header with key "file" and
	// value "testimage.png" in the form, save that to part
	part, err := writer.CreateFormFile("file", "testimage.png")
	if err != nil {
		t.Fatal(err)
	}

	// copy fileToUpload's data into part
	_, err = io.Copy(part, fileToUpload)
	if err != nil {
		t.Fatal("Failed to copy file contents:", err)
	}

	writer.Close()

	// build a new request
	req, err := http.NewRequest(http.MethodPost, lc+"/upload", &requestBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// create a recorder for the response
	rr := httptest.NewRecorder()

	// send the request, save response to recorder
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("POST to '/upload' -> expected code 200, got code", rr.Code)
	}

	expected := "File uploaded successfully"
	if !bytes.Contains(rr.Body.Bytes(), []byte(expected)) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}

	// check if file is in DB
	filename := filepath.Base(pathToTestImage)
	query := "SELECT filename FROM uploads WHERE filename = $1"
	row := app.Connection.QueryRow(context.TODO(), query, filename)
	var dbFile string
	row.Scan(&dbFile)
	if dbFile != filename {
		t.Error("couldn't find file in DB")
	}

	// check if file is in uploadDir
	files, err := os.ReadDir(app.UploadDir)
	if err != nil {
		t.Error(err)
	}
	for _, fileName := range files {
		if fileName.Name() == filename {
			return
		}
	}
	t.Error("couldn't find testfile in app.UploadDir")

}

func TestHandleGetAllFiles(t *testing.T) {
	app := &Config{
		Connection:        connectToDBTest(), // separate connection to db for tests
		UploadDir:         "../uploads/",     // since not in docker, dir is different
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}
	r := app.routes()
	defer app.Connection.Close(context.Background())

	req, err := http.NewRequest(http.MethodGet, lc+"/files", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var response jsonResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Failed to unmarshal response body:", err)
	}

	if response.Status != http.StatusOK {
		t.Error("GET to '/files' -> expected code 200, got code", rr.Code)
	}

	expected := "testimage.png"
	if !strings.Contains(response.Message, expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleGetFileByID(t *testing.T) {
	app := &Config{
		Connection:        connectToDBTest(), // separate connection to db for tests
		UploadDir:         "../uploads/",     // since not in docker, dir is different
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}
	r := app.routes()
	defer app.Connection.Close(context.Background())

	var randID string
	query := "SELECT id FROM uploads ORDER BY random() LIMIT 1;"
	row := app.Connection.QueryRow(context.TODO(), query)
	row.Scan(&randID)

	req, err := http.NewRequest(http.MethodGet, lc+"/files/"+randID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("GET to '/files/<random>' -> expected code 200, got code", rr.Code)
	}

	expected := "file found"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleDeleteFileByID(t *testing.T) {
	app := &Config{
		Connection:        connectToDBTest(), // separate connection to db for tests
		UploadDir:         "../uploads/",     // since not in docker, dir is different
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}
	r := app.routes()
	defer app.Connection.Close(context.Background())

	initStat, err := os.Stat(app.UploadDir)
	if err != nil {
		t.Fatal(err)
	}

	var randID string
	query := "SELECT id FROM uploads ORDER BY random() LIMIT 1;"
	row := app.Connection.QueryRow(context.TODO(), query)
	row.Scan(&randID)

	req, err := http.NewRequest(http.MethodDelete, lc+"/files/"+randID, nil)
	if err != nil {
		t.Fatal("Couldn't build request", err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var response jsonResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Couldn't unmarshal into response")
	}

	if response.Status != 200 {
		t.Error("Expected status 200, got", response.Status)
	}

	expected := "successfully"
	if !strings.Contains(response.Message, expected) {
		t.Errorf("Expected %s in Message, got %s", expected, response.Message)
	}

	newStat, err := os.Stat(app.UploadDir)
	if err != nil {
		t.Error(err)
	}

	if initStat == newStat {
		t.Error("File wasn't deleted")
	}
}

func connectToDBTest() *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@0.0.0.0:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_USER"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 1; i <= 10; i++ {
		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			log.Println("postgres not yet ready...")
			log.Println(err)
		} else {
			log.Println("Connected to database!")
			time.Sleep(1 * time.Second)

			err = conn.Ping(ctx)
			if err != nil {
				log.Fatal("Failed to ping DB!")
				conn.Close(ctx)
			}
			log.Println("Ping successful!")

			log.Println("Established DB connection!")
			return conn
		}

		log.Println("backing off for 1 second...")
		time.Sleep(1 * time.Second)
	}
	return nil
}
