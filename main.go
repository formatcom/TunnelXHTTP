package main

import (
	"io"
	"log"
	"fmt"
	"flag"
	"net/http"
	"net/http/httputil"
)

var (
	listen = flag.String ("listen", ":8000", "")
	tls    = flag.Bool   ("tls",      false, "")
)

func main() {

	flag.Parse()

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		_req, _ := httputil.DumpRequest(req, true)

		fmt.Println(string(_req))

		// proxy here .
		io.WriteString(w, "Hello, world!\n")
	}

	// capture all requests
	http.HandleFunc("/", helloHandler)

	if *tls {
		log.Fatal(http.ListenAndServeTLS(*listen,
				"key/server.pem", "key/server.key.pem", nil))
	} else {
		log.Fatal(http.ListenAndServe(*listen, nil))
	}
}
