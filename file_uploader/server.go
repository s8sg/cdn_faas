package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	jsonType   = "application/json"
	tmpDIR     = "/home/app/"
	storageURL = "http://file-storage:8080"
)

var (
	uploadQueue = make(chan string, 10)
	failedQueue = make(chan string, 10)
)

// upload logic
func uploadHandle(w http.ResponseWriter, r *http.Request) {

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
	filesize := header.Size

	log.Printf("received upload request for file '%s' with size '%d'", filename, filesize)

	// Save the file locally
	filePath := tmpDIR + filename
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
	fi, err := f.Stat()
	if err != nil {
		log.Printf("failed to get file stat, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
		f.Close()
		return
	} else {
		log.Printf("file saved on the temp directory as %s filesize %d", filePath, fi.Size())
		f.Close()
	}

	// Put the filename to upload channel
	uploadQueue <- filename

	w.Write([]byte(fmt.Sprintf("File successfully queued for upload %s", filePath)))
}

// file upload logic
func upload(client *http.Client, url string, filename string, r io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer

	if x, ok := r.(io.Closer); ok {
		defer x.Close()
	}
	// Add an image file
	if fw, err = w.CreateFormFile("file", filename); err != nil {
		return
	}
	if _, err = io.Copy(fw, r); err != nil {
		return err
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

// upload file to the storage
func uploadToStorage(filename string) error {
	filePath := tmpDIR + filename
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("failed to read file locally, error %v", err)
		return fmt.Errorf("failed to read file locally, error %v", err)
	}
	defer f.Close()
	client := &http.Client{}
	err = upload(client, storageURL, filename, f)
	if err != nil {
		log.Printf("failed send image to storage, error %v", err)
		return err
	}
	return nil
}

// remove temp file
func remove(fileName string) {
	log.Printf("deleting file '%s'", fileName)
	filePath := tmpDIR + fileName
	os.Remove(filePath)
}

// The uploader thread
func uploader() {
	for true {
		// read from channel
		select {
		case fileName := <-uploadQueue:
			log.Printf("New file '%s' received to upload", fileName)
			err := uploadToStorage(fileName)
			if err != nil {
				log.Printf("failed to handle file '%s' upload, error %v", fileName, err)
				log.Printf("failed file '%'s will be queued for retry", fileName)
				failedQueue <- fileName
			} else {
				log.Printf("successfully uploaded file '%s' to storage", fileName)
				//remove(fileName)
			}
		case fileName := <-failedQueue:
			log.Printf("Failed file '%s' received to retry", fileName)
			err := uploadToStorage(fileName)
			if err != nil {
				log.Printf("failed to upload file '%s' in retry, error %v", fileName, err)
				remove(fileName)
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func main() {

	go uploader()

	http.HandleFunc("/", uploadHandle)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
