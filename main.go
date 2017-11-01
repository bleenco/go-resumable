package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

func main() {
	isServer := flag.Bool("server", false, "")
	isClient := flag.Bool("client", false, "")
	flag.Parse()

	if *isServer {
		server()
	}

	if *isClient {
		url := "http://localhost:2110"
		filePath := "/Users/jan/Desktop/out.dmg"

		client := New()
		err := client.Upload(url, filePath)
		if err != nil {
			panic(err)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sample() {
	filePath := "/Users/jan/Desktop/test-repo/ubuntu-17.10-desktop-amd64.iso"

	file, err := os.Open(filePath)
	checkError(err)
	defer file.Close()

	fileInfo, _ := file.Stat()
	var fileSize = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1MB

	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)

		// write to disk
		fileName := "/Users/jan/Desktop/pieces/somebigfile_" + strconv.FormatUint(i, 10)
		_, err := os.Create(fileName)
		checkError(err)

		// write/save buffer to disk
		ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)
		fmt.Println("Split to:", fileName)
	}

	// recombine back the chunked files in a new file
	newFile := "/Users/jan/Desktop/ubuntu.iso"
	_, err = os.Create(newFile)
	checkError(err)

	file, err = os.OpenFile(newFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	checkError(err)

	var writePosition int64

	for j := uint64(0); j < totalPartsNum; j++ {
		currentChunkFile := "/Users/jan/Desktop/pieces/somebigfile_" + strconv.FormatUint(j, 10)
		newFileChunk, err := os.Open(currentChunkFile)
		checkError(err)

		defer newFileChunk.Close()
		chunkInfo, err := newFileChunk.Stat()
		checkError(err)

		var chunkSize = chunkInfo.Size()
		chunkBufferBytes := make([]byte, chunkSize)

		fmt.Println("Appending at position: [", writePosition, "] bytes")
		writePosition = writePosition + chunkSize

		reader := bufio.NewReader(newFileChunk)
		_, err = reader.Read(chunkBufferBytes)
		checkError(err)

		n, err := file.Write(chunkBufferBytes)
		checkError(err)

		file.Sync()

		chunkBufferBytes = nil
		fmt.Println("Written", n, "bytes")
		fmt.Println("Recombining part [", j, "] info:", newFile)
	}

	file.Close()
}
