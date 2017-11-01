# go-resumable

It's a Go library providing multiple simultaneous and resumable uploads.

Library is designed to introduce fault-tolerance into the upload of large files throught HTTP.
This is done by splitting each file into small chunks; whenever the upload of a chunk fails, uploading is retried until the procedure completes.
This allows uploads to automatically resume uploading after a network connection is lost either locally or to the server.
Additionally, it allows users to pause, resume and even recover uploads without losing state.

### Installation

```sh
$ go get -v https://github.com/bleenco/go-resumable
```

### Usage

```go
import (
  "net/http"

  "github.com/bleenco/go-resumable"
)

func main() {
  httpClient := &http.Client{}
  url := "http://example.com/upload"
  filePath := "/path/to/file/to/upload.zip"
  chunkSize := int(1 * (1 << 20)) // 1MB
  client := resumable.New(url, filePath, httpClient, chunkSize)

  client.Init()
  client.Start()

  resumable.WG.Wait() // this is important
}
```

### Licence

MIT
