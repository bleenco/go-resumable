package main

import (
	"crypto/rand"
	"fmt"
)

func fatalErrCheck(err error) {
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
}

func generateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%X", b)
}

func generateContentRange(index uint64, fileChunk int, partSize int, totalSize int64) string {
	var contentRange string
	if index == 0 {
		contentRange = "bytes 0-" + fmt.Sprintf("%v", partSize) + "/" + fmt.Sprintf("%v", totalSize)
	} else {
		from := uint64(fileChunk) * index
		to := uint64(fileChunk) * (index + 1)
		contentRange = "bytes " + fmt.Sprintf("%v", from) + "-" + fmt.Sprintf("%v", to) + "/" + fmt.Sprintf("%v", totalSize)
	}

	return contentRange
}
