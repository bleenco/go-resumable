package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/bleenco/go-resumable"
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
		filePath := "/Users/jan/Desktop/ubuntu-17.10-desktop-amd64.iso"
		file2 := "/Users/jan/Desktop/ubuntu-17.10-desktop-amd64.iso.zip"

		httpClient := &http.Client{}
		chunkSize := int(1 * (1 << 20)) // 1MB
		// chunkSize := 10000

		client := resumable.New(url, filePath, httpClient, chunkSize, true)
		client2 := resumable.New(url, file2, httpClient, chunkSize, true)

		client.Init()
		client.Start()
		client2.Init()
		client2.Start()

		time.Sleep(1 * time.Second)

		client.Pause()
		client2.Pause()

		fmt.Println("Already transferred (iso):", client.Status.SizeTransferred, "/", client.Status.Size)
		fmt.Println("Already transferred (zip):", client2.Status.SizeTransferred, "/", client2.Status.Size)
		time.Sleep(2 * time.Second)
		client.Start()
		client2.Start()

		resumable.WG.Wait()
	}
}
