package resumable

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
)

// Resumable structure
type Resumable struct {
	client    *http.Client
	url       string
	filePath  string
	id        string
	chunkSize int
}

// UploadStatus holds the data about upload
type UploadStatus struct {
	size            int64
	transfered      int64
	parts           uint64
	partsTransfered uint64
}

// Uploads holds the uploads progresses
var Uploads = make(map[string]UploadStatus)

// New creates new instance of resumable Client
func New(url string, filePath string, client *http.Client, chunkSize int) *Resumable {
	resumable := &Resumable{
		client:    client,
		url:       url,
		filePath:  filePath,
		id:        generateSessionID(),
		chunkSize: chunkSize,
	}

	return resumable
}

// StartUpload initializes upload
func (c *Resumable) StartUpload() error {
	file, err := os.Open(c.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	var totalSize = fileStat.Size()
	totalPartsNum := uint64(math.Ceil(float64(totalSize) / float64(c.chunkSize)))

	var uploadStatus UploadStatus

	_, ok := Uploads[c.id]
	if !ok {
		Uploads[c.id] = UploadStatus{
			size:            totalSize,
			transfered:      0,
			parts:           totalPartsNum,
			partsTransfered: 0,
		}
	}

	uploadStatus = Uploads[c.id]

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(float64(c.chunkSize), float64(totalSize-int64(i*uint64(c.chunkSize)))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)
		contentRange := generateContentRange(i, c.chunkSize, partSize, totalSize)

		responseBody, err := httpRequest(c.url, c.client, c.id, totalSize, partBuffer, contentRange)
		if err != nil {
			return err
		}

		uploadStatus.transfered = parseBody(responseBody)
		uploadStatus.partsTransfered = i + 1

		fmt.Println(uploadStatus)
	}

	return nil
}

// Pause upload by sessionID
// func (c *Resumable) Pause(sessionID string) error {

// }

func httpRequest(url string, client *http.Client, sessionID string, totalSize int64, part []byte, contentRange string) (string, error) {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(part))
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", "application/octet-stream")
	request.Header.Add("Content-Disposition", "attachment; filename='out.dmg'")
	request.Header.Add("Content-Range", contentRange)
	request.Header.Add("Session-ID", sessionID)

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
