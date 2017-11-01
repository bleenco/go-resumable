package resumable

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"time"
)

type uploadFile struct {
	file       *os.File
	name       string
	tempPath   string
	status     string
	size       int64
	transfered int64
}

var files = make(map[string]uploadFile)

type fileStorage struct {
	Path     string
	TempPath string
}

// FileStorage settings.
// When finished uploading with success files are stored inside Path config.
// While uploading temporary files are stored inside TempPath directory.
var FileStorage = fileStorage{
	Path:     "./files",
	TempPath: ".tmp",
}

// HTTPHandler is main request/response handler for HTTP server.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	ensureDir(FileStorage.Path)
	ensureDir(FileStorage.TempPath)

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

			_, params, err := mime.ParseMediaType(r.Header.Get("Content-Disposition"))
			checkError(err)
			fileName := params["filename"]

			newFile := FileStorage.TempPath + "/" + sessionID
			_, err = os.Create(newFile)
			checkError(err)

			f, err := os.OpenFile(newFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			checkError(err)

			files[sessionID] = uploadFile{
				file:     f,
				name:     fileName,
				tempPath: newFile,
				status:   "created",
				size:     totalSize,
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
		moveToPath(sessionID)
		upload.file.Close()
		delete(files, sessionID)
	}
}

func moveToPath(id string) {
	uploadFile := files[id]
	filePath := FileStorage.Path + "/" + uploadFile.name
	if fileExists(filePath) {
		t := time.Now().Format(time.RFC3339)
		filePath = FileStorage.Path + "/" + t + "-" + uploadFile.name
	}

	err := os.Rename(uploadFile.tempPath, filePath)
	checkError(err)
}
