package resumable

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"runtime"
	"sync"
)

const (
	stopped = 0
	paused  = 1
	running = 2
)

// WG exports wait group, so we can wait for it
var WG sync.WaitGroup

// Resumable structure
type Resumable struct {
	client    *http.Client
	url       string
	filePath  string
	id        string
	chunkSize int
	file      *os.File
	channel   chan int
	Status    UploadStatus
}

// UploadStatus holds the data about upload
type UploadStatus struct {
	Size             int64
	SizeTransferred  int64
	Parts            uint64
	PartsTransferred uint64
}

// New creates new instance of resumable Client
func New(url string, filePath string, client *http.Client, chunkSize int) *Resumable {
	resumable := &Resumable{
		client:    client,
		url:       url,
		filePath:  filePath,
		id:        generateSessionID(),
		chunkSize: chunkSize,
		Status: UploadStatus{
			Size:             0,
			SizeTransferred:  0,
			Parts:            0,
			PartsTransferred: 0,
		},
	}

	return resumable
}

// Init method initializes upload
func (c *Resumable) Init() {
	fileStat, err := os.Stat(c.filePath)
	checkError(err)

	c.Status.Size = fileStat.Size()
	c.Status.Parts = uint64(math.Ceil(float64(c.Status.Size) / float64(c.chunkSize)))

	c.channel = make(chan int, 1)
	c.file, err = os.Open(c.filePath)
	checkError(err)
	defer c.file.Close()
	WG.Add(1)

	go func() {
		c.upload()
		c = nil
		WG.Done()
	}()
}

// Start set upload state to uploading
func (c *Resumable) Start() {
	c.channel <- 2
}

// Pause set upload state to paused
func (c *Resumable) Pause() {
	c.channel <- 1
}

// Cancel set upload state to stopped
func (c *Resumable) Cancel() {
	c.channel <- 0
}

func (c *Resumable) upload() {
	state := paused
	i := uint64(0)

	for {
		select {
		case state = <-c.channel:
			switch state {
			case stopped:
				fmt.Printf("Worker %s: stopped\n", c.id)
				return
			case running:
				fmt.Printf("Worker %s: running\n", c.id)
			case paused:
				fmt.Printf("Worker %s: paused\n", c.id)
			}

		default:
			runtime.Gosched()
			if state == paused {
				break
			}

			c.uploadChunk(i)
			i = i + 1
		}
	}
}

func (c *Resumable) uploadChunk(i uint64) {
	if i+1 == c.Status.Parts {
		WG.Done()
	} else {
		partSize := int(math.Min(float64(c.chunkSize), float64(c.Status.Size-int64(i*uint64(c.chunkSize)))))
		partBuffer := make([]byte, partSize)
		c.file.Read(partBuffer)
		contentRange := generateContentRange(i, c.chunkSize, partSize, c.Status.Size)

		responseBody, err := httpRequest(c.url, c.client, c.id, c.Status.Size, partBuffer, contentRange)
		checkError(err)

		c.Status.SizeTransferred = parseBody(responseBody)
		c.Status.PartsTransferred = i + 1
	}
}

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
