package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func server() {
	http.HandleFunc("/", HTTPHandler)
	fmt.Println("Listening on http://localhost:2110")
	http.ListenAndServe(":2110", nil)
}

var file *os.File

// HTTPHandler is main request/response handler for HTTP server.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Session-ID") != "" && r.Header.Get("Content-Range") != "" {
		sessionID := r.Header.Get("Session-ID")
		contentRange := r.Header.Get("Content-Range")

		body, err := ioutil.ReadAll(r.Body)
		checkError(err)

		contentRange = strings.Replace(contentRange, "bytes ", "", -1)
		fromTo := strings.Split(contentRange, "/")[0]
		totalSize, _ := strconv.ParseInt(strings.Split(contentRange, "/")[1], 10, 64)
		splitted := strings.Split(fromTo, "-")
		partFrom, _ := strconv.ParseInt(splitted[0], 10, 64)
		partTo, _ := strconv.ParseInt(splitted[1], 10, 64)

		if partFrom == 0 {
			newFile := "/Users/jan/Desktop/test-data/" + sessionID + ".dmg"
			_, err = os.Create(newFile)
			checkError(err)

			file, err = os.OpenFile(newFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			checkError(err)
		}

		_, err = file.Write(body)
		checkError(err)

		file.Sync()

		if partFrom == 0 {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Header().Set("Content-Length", "100")
		w.Header().Set("Connection", "close")
		w.Header().Set("Range", contentRange)
		w.Write([]byte(contentRange))

		if partTo >= totalSize {
			file.Close()
		}
	}
}
