package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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

		httpClient := &http.Client{}
		chunkSize := int(1 * (1 << 20)) // 1MB
		client := New(url, filePath, httpClient, chunkSize)
		err := client.StartUpload()
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
