package main

import (
	"log"
	"os"
	"slices"
	"strings"
)

// Checks if upload dir exists, if not, creates it
func (app *Config) checkUploadDirExists() {
	stat, err := os.Stat(app.UploadDir)

	if err == nil && stat.IsDir() {
		return
	} else {
		err := os.MkdirAll(app.UploadDir, os.ModePerm)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

// Returns the file extension from the filename
func (app *Config) getFileExtension(file string) string {
	extension := strings.LastIndex(file, ".")
	if extension == -1 {
		return ""
	}

	return file[extension+1:]
}

// Checks if entry is in app.AllowedExtensions
func (app *Config) isValidImageExtension(entry string) bool {
	if !slices.Contains(app.AllowedExtensions, entry) {
		return false
	} else {
		return true
	}
}
