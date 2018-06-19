package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	jsonType = "application/json"
)

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {

	defer w.Header().Set("Content-Type", jsonType)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't get fie from request\"}", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)
}

func main() {

	http.HandleFunc("/", upload)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
