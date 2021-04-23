testserv - Control a test HTTP server with simple instructions
==============================================================

Getting Started
===============

Install testserv:

```sh
$ go get github.com/gogama/testserv
```

Import the testserv package and create a test HTTP server to start testing!

```go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogama/testserv"
)

func TestMyThing(t *testing.T) {
	server := httptest.NewServer(&testserv.Handler{
		Inst: []testserv.Instruction{
			// Server will delay half a second, then return a 200 OK response
			// containing the given body.
			{
				HeaderDelay: 500*time.Millisecond,
				StatusCode: 200,
				Body: []byte(`{}`),
			},
		},
	})
	client := server.Client()

	resp, err := client.Get(server.URL)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %s\n", err)
		return
	}

	fmt.Printf("StatusCode: %d\n", resp.StatusCode)
	// 200
	fmt.Printf("Body: %s\n", body)
	// {}
}
```
