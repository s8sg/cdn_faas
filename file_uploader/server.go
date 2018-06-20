package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	jsonType = "application/json"
	homeDIR  = "/home/app/"
)

var (
	validType = map[string]bool{
		".png": true,
	}

	uploadQueue = make(chan string, 10)
	failedQueue = make(chan string, 10)
)

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {

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
	filePath := homeDIR + filename
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
		return
	}
	io.Copy(f, r.Body)
	log.Printf("Resized file saved on the local directory as %s", filePath)
	f.Close()

	// Put the filename to upload channel
	uploadQueue <- filename

	w.Write([]byte(fmt.Sprintf("File successfully saved on the local directory as %s", filePath)))
}

func uploadRemote(filename string) error {
	log.Printf("unimplemented: file uplaoded successfully")
	return nil
}

// The uploader thread
func uploader() {
	for true {

		select {
		// read from channel
		case fileName := <-uploadQueue:
			log.Printf("Uploader received a file %s to upload", fileName)
			err := uploadRemote(fileName)
			if err != nil {
				log.Printf("Failed to handle file upload will retry")
				failedQueue <- fileName
				time.Sleep(100 * time.Millisecond)
			}
		case fileName := <-failedQueue:
			log.Printf("Uploader received a file %s that was failed before", fileName)
			err := uploadRemote(fileName)
			if err != nil {
				log.Printf("Failed to upload file, deleting file %s", fileName)
				filePath := homeDIR + fileName
				os.Remove(filePath)
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func main() {

	go uploader()

	http.HandleFunc("/", upload)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
