package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/jkuri/go-resumable"
)

func main() {
	isServer := flag.Bool("server", false, "")
	isClient := flag.Bool("client", false, "")
	flag.Parse()

	if *isServer {
		http.HandleFunc("/", resumable.HTTPHandler)
		fmt.Println("Listening on http://localhost:2110")
		http.ListenAndServe(":2110", nil)
	}

	if *isClient {
		url := "http://localhost:2110"
		filePath := "/Users/jan/Desktop/out.dmg"

		httpClient := &http.Client{}
		// chunkSize := int(1 * (1 << 20)) // 1MB
		chunkSize := 10000
		client := resumable.New(url, filePath, httpClient, chunkSize)
		err := client.StartUpload()
		if err != nil {
			panic(err)
		}
	}
}
