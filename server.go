package resumable

import (
	"io/ioutil"
	"net/http"
	"os"
)

type uploadFile struct {
	file       *os.File
	status     string
	size       int64
	transfered int64
}

var files = make(map[string]uploadFile)

// HTTPHandler is main request/response handler for HTTP server.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("Session-ID") == "" || r.Header.Get("Content-Range") == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Invalid request."))
	}

	var upload uploadFile

	sessionID := r.Header.Get("Session-ID")
	contentRange := r.Header.Get("Content-Range")

	body, err := ioutil.ReadAll(r.Body)
	checkError(err)

	totalSize, partFrom, partTo := parseContentRange(contentRange)

	if partFrom == 0 {
		_, ok := files[sessionID]
		if !ok {
			w.WriteHeader(http.StatusCreated)
			newFile := "/Users/jan/Desktop/test-data/" + sessionID + ".dmg"
			_, err = os.Create(newFile)
			checkError(err)

			f, err := os.OpenFile(newFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			checkError(err)

			files[sessionID] = uploadFile{
				file:   f,
				status: "created",
				size:   totalSize,
			}
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	upload = files[sessionID]
	upload.status = "uploading"

	_, err = upload.file.Write(body)
	checkError(err)

	upload.file.Sync()
	upload.transfered = partTo

	w.Header().Set("Content-Length", string(len(body)))
	w.Header().Set("Connection", "close")
	w.Header().Set("Range", contentRange)
	w.Write([]byte(contentRange))

	if partTo >= totalSize {
		upload.file.Close()
		delete(files, sessionID)
	}
}
