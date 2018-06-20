package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	jsonType   = "application/json"
	resizerURL = "http://image-resizer:8080"
	homeDIR    = "/home/app/"
)

var (
	validType = map[string]bool{
		".png": true,
	}
)

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {

	defer w.Header().Set("Content-Type", jsonType)

	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("uploadfile")
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't get file from request\"}", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := header.Filename
	extn := filepath.Ext(filename)

	if !validType[extn] {
		log.Printf("invalid file type %s", extn)
		http.Error(w, "Invalid file type, only .png is supported for now", http.StatusInternalServerError)
		return
	}

	// Read the file
	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("failed to read file %s from request, error %v", filename, extn)
		http.Error(w, fmt.Sprintf("failed to read file %s from request", filename), http.StatusInternalServerError)
		return
	}

	// Try to Resize the file
	client := &http.Client{}
	req, _ := http.NewRequest("POST", resizerURL, bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Accept", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to requet resizer for file %s, error %v", filename, err)
		http.Error(w, fmt.Sprintf("failed to resize file %s", filename), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("failed to requet resizer for file %s, response code %d", filename, resp.StatusCode)
		http.Error(w, fmt.Sprintf("failed to resize file %s", filename), http.StatusInternalServerError)
		return
	}

	/*
		resizedData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read resized file %s, error %v", filename, resizedData)
			http.Error(w, fmt.Sprintf("failed to resize file %s", filename), http.StatusInternalServerError)
			return
		}*/

	// Save the file locally
	filePath := homeDIR + filename
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("failed to upload, error %v", err)
		http.Error(w, "{\"error\":\"Couldn't write file\"}", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, resp.Body)
	log.Printf("Resized file saved on the local directory as %s", filePath)

	w.Write([]byte(fmt.Sprintf("File successfully saved on the local directory as %s", filePath)))
}

func main() {

	http.HandleFunc("/", upload)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
