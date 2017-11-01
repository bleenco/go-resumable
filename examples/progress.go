package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/bleenco/go-resumable"
	pb "gopkg.in/cheggaaa/pb.v1"
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

		httpClient := &http.Client{}
		chunkSize := int(1 * (1 << 20)) // 1MB

		client := resumable.New(url, filePath, httpClient, chunkSize, false)

		client.Init()

		count := client.Status.Size
		bar := pb.New(int(count))
		bar.SetRefreshRate(time.Millisecond)
		bar.SetUnits(pb.U_BYTES)

		go func() {
			bar.Start()
			for {
				if client.Status.SizeTransferred > client.Status.Size {
					bar.FinishPrint("Done.")
					break
				}
				bar.Set(int(client.Status.SizeTransferred))
				time.Sleep(time.Millisecond)
			}
		}()

		client.Start()
		time.Sleep(1 * time.Second)
		client.Pause()
		time.Sleep(2 * time.Second)
		client.Start()

		resumable.WG.Wait()
	}
}
