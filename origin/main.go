// A tiny web server that serves a static file.
package main

import (
	//"flag"
	"fmt"
	//"log"
	"net/http"
)

func main() {

	fs := http.StripPrefix("/", http.FileServer(http.Dir("/html/")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private")
		// Needed for local proxy to Kubernetes API server to work.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,Cache-Control,Content-Type")
		// Disable If-Modified-Since so update-demo isn't broken by 304s
		r.Header.Del("If-Modified-Since")
		fs.ServeHTTP(w, r)
	})

	fmt.Println("Started, serving at 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

	select {}
}
