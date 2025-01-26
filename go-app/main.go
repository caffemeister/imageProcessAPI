package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
)

var uploadDir = "./uploads/"
var maxFileSize = 10 << 20                                                      // ~10 mb
var allowedExtensions = []string{"png", "jpg", "jpeg"}                          // allowed types of files
var usageInfo = "POST: /uploads\nGET: /files, /files/<id>\nDELETE: /files/<id>" // for GET to "/", shows usage

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &Config{
		Connection:        connectToDB(),
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}

	// check for db table, create if not exist
	// app.checkDBTable()

	// check for ./uploads/
	app.checkUploadDirExists()

	// create a new router
	r := app.routes()

	// channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// set up server
	server := &http.Server{
		Addr:    ":8001",
		Handler: r,
	}

	// launch server on separate goroutine
	go func() {
		log.Println("Starting...")
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// blocker until receive stop signal from system
	<-stop
	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shut down forcefully: %v", err)
	}

	log.Println("Closing DB connection...")
	if err := app.Connection.Close(shutdownCtx); err != nil {
		log.Fatalf("Failed to close connection to DB: %v", err)
	}

	log.Println("Shutdown successful! Bye.")
}

func connectToDB() *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@postgres:5432/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_USER"))

	for i := 1; i <= 10; i++ {
		conn, err := pgx.Connect(context.Background(), connStr)
		if err != nil {
			log.Println("postgres not yet ready...")
			log.Println(err)
		} else {
			log.Println("Connected to database!")

			err = conn.Ping(context.Background())
			if err != nil {
				log.Fatal("Failed to ping DB!")
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

func (app *Config) checkDBTable() {
	query := "SELECT to_regclass('public.uploads');"

	var tableExists string
	err := app.Connection.QueryRow(context.Background(), query).Scan(&tableExists)
	if err != nil {
		log.Fatalf("error checking if table exists: %v", err)
	}

	if tableExists == "" {
		createTableQuery := `
			CREATE TABLE public.uploads (
				id SERIAL PRIMARY KEY,
				filename VARCHAR(255) NOT NULL,
				filepath TEXT NOT NULL,
				uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`

		_, err := app.Connection.Exec(context.Background(), createTableQuery)
		if err != nil {
			log.Fatalf("Error creating table: %v", err)
		}
		fmt.Println("UPLOADS table created!")
	} else {
		fmt.Println("UPLOADS table already exists.")
	}
}
