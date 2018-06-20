package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

const (
	jsonType    = "application/json"
	resizerURL  = "http://image-resizer:8080"
	uploaderURL = "http://file-uploader:8080"
	homeDIR     = "/home/app/"
)

var (
	validType = map[string]bool{
		".png": true,
	}
)

// file upload logic
func Upload(client *http.Client, url string, filename string, r io.Reader) (err error) {
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

// file handle logic
func handle(w http.ResponseWriter, r *http.Request) {

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

	// send the file to file uploader
	client = &http.Client{}
	err = Upload(client, uploaderURL, filename, resp.Body)
	if err != nil {
		log.Printf("failed to send resized image to uploader, error %v", err)
		http.Error(w, fmt.Sprintf("failed to upload file %s", filename), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Resized file successfully uploaded to uploader"))
}

func main() {

	http.HandleFunc("/", handle)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
