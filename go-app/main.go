package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

var uploadDir = "./../uploads/"
var maxFileSize = 10 << 20                                 // ~10 mb
var allowedExtensions = []string{"png", "jpg", "jpeg"}     // allowed types of files
var usageInfo = "POST: /uploads\nGET: /files, /files/<id>" // for GET to "/", shows usage

type Config struct {
	Connection        *pgx.Conn
	UploadDir         string
	MaxFileSize       int
	AllowedExtensions []string
}

type jsonResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	File    string `json:"file,omitempty"`
}

func main() {
	app := &Config{
		Connection:        connectToDB(),
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}

	// check for ./uploads/
	app.checkUploadDirExists()

	// create a new router
	r := app.routes()

	http.ListenAndServe(":8080", r)
	defer app.Connection.Close(context.Background())
}

func connectToDB() *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_USER"))

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to ping DB!")
	}
	log.Println("Ping successful!")

	log.Println("Established DB connection!")
	return conn
}
