package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
)

var (
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

// Part upload part structure
type Part struct {
	url       string
	path      string
	rangeFrom int64
	rangeTo   int64
}

// Request structure
type Request struct {
	url       string
	resumable bool
	file      string
	par       int64
	len       int64
	parts     []Part
}

// Resumable structure
type Resumable struct {
	uploads []Request
}

// New creates new instance of resumable Client
func New() *Resumable {
	var requests []Request
	client := &Resumable{
		uploads: requests,
	}

	return client
}

// Upload initializes upload
func (c *Resumable) Upload(url string, filePath string) error {
	id := generateSessionID()
	client := &http.Client{Transport: transport}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	var totalSize = fileStat.Size()
	const fileChunk = 1 * (1 << 20) // 1MB
	totalPartsNum := uint64(math.Ceil(float64(totalSize) / float64(fileChunk)))

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(totalSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)
		index := uint64(i)

		var contentRange string
		if index == 0 {
			contentRange = "bytes 0-" + fmt.Sprintf("%v", partSize) + "/" + fmt.Sprintf("%v", totalSize)
		} else {
			from := fileChunk * i
			to := fileChunk * (i + 1)
			contentRange = "bytes " + fmt.Sprintf("%v", from) + "-" + fmt.Sprintf("%v", to) + "/" + fmt.Sprintf("%v", totalSize)
		}

		err := c.Request(url, client, id, totalSize, index, partBuffer, contentRange)
		if err != nil {
			return err
		}
	}

	return nil
}

// Request initializes HTTP request
func (c *Resumable) Request(url string, client *http.Client, sessionID string, totalSize int64, index uint64, part []byte, contentRange string) error {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(part))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/octet-stream")
	request.Header.Add("Content-Disposition", "attachment; filename='out.dmg'")
	request.Header.Add("Content-Range", contentRange)
	request.Header.Add("Session-ID", sessionID)

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fmt.Println("Status:", response.Status)
	fmt.Println("Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("Body:", string(body))

	return nil
}
