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

	dbConn := connectToDB()
	if dbConn == nil {
		log.Fatal("couldn't connect to DB")
	}

	app := &Config{
		Connection:        dbConn,
		UploadDir:         uploadDir,
		MaxFileSize:       maxFileSize,
		AllowedExtensions: allowedExtensions,
	}

	// check for ./uploads/
	app.checkUploadDirExists()

	// channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// set up server
	server := &http.Server{
		Addr:    ":8001",
		Handler: app.routes(),
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
	if app.Connection != nil {
		if err := app.Connection.Close(shutdownCtx); err != nil {
			log.Fatalf("Failed to close connection to DB: %v", err)
		}
	}

	log.Println("Shutdown successful! Bye.")
}

func connectToDB() *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@postgres:5432/%s?sslmode=disable",
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
