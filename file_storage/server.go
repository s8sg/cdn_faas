package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	jsonType   = "application/json"
	storageDIR = "/home/app/storage/"
)

// Static file request handler
func sendFile(w http.ResponseWriter, r *http.Request, file string) {
	log.Printf("serving file %s", file)
	filepath := storageDIR + file
	http.ServeFile(w, r, filepath)
}

// storage handler
func storageHandle(w http.ResponseWriter, r *http.Request) {

	// Check request type
	switch r.Method {
	case "GET":
		file := r.URL.Query().Get("file")
		if len(file) > 0 {
			sendFile(w, r, file)
		}

	case "POST":
		defer w.Header().Set("Content-Type", jsonType)

		r.ParseMultipartForm(32 << 20)
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("failed to upload, error %v", err)
			http.Error(w, "{\"error\":\"Couldn't get file from request\"}", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		filename := header.Filename

		// Save the file locally
		filePath := storageDIR + filename
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("failed to upload, error %v", err)
			http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
			return
		}
		io.Copy(f, file)
		log.Printf("File successfully saved, available at 'function/file-storage?file=%s'", filename)
		f.Close()

		w.Write([]byte(fmt.Sprintf("File successfully saved, available at 'function/file-storage?file=%s'", filename)))
	default:
		log.Printf("Invalid request %s", r.Method)
	}
}

func main() {
	http.HandleFunc("/", storageHandle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
