package main

import (
	"os"
	"fmt"
	"flag"
	"bytes"
	"net"
	"strconv"
	"errors"
	"bufio"
	"strings"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/http/httputil"
)

var (
	listen  = flag.String ("listen", ":8000", "")
	withTLS = flag.Bool   ("tls",      false, "")
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

/* Para mi
transport := &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

client := &http.Client{Transport: transport}

resp, err := client.Get("https://google.com")
*/

func main() {

	flag.Parse()

	var err error
	var server net.Listener

	if *withTLS {

		cert, err := tls.LoadX509KeyPair("key/server.pem", "key/server.key.pem")
		checkError(err)

		conf := &tls.Config{
				InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert},
		}

		server, err = tls.Listen("tcp", *listen, conf)
		checkError(err)

	} else {
		server, err = net.Listen("tcp", *listen)
		checkError(err)
	}

	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil { continue }

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, dat, err := readRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-\n%s-\n", dat)

	dump, _ := httputil.DumpRequest(req, false)

	fmt.Printf("\n-\n%s-\n", dump)


	response := http.Response{
				StatusCode:    200,
				ProtoMajor:    req.ProtoMajor,
				ProtoMinor:    req.ProtoMinor,
				Request:       req,
				Header:        http.Header{},
				Body:          ioutil.NopCloser(bytes.NewReader(dat[:])),
				ContentLength: -1,
	}

	var braw bytes.Buffer
	response.Write(&braw)

	resp := braw.Bytes()

	fmt.Println(string(resp))

	// proxy here .
	conn.Write(resp)

}

func readRequest(b *bufio.Reader) (req *http.Request, data []byte, err error) {

	tp  := textproto.NewReader(b)
	req  = new(http.Request)

	req.Header = http.Header{}

	var s string

	// First line: GET /index.html HTTP/1.0
	if s, err = tp.ReadLine(); err != nil {
		return nil, nil, err
	}

	data = append(data, []byte(s)...)
	data = append(data, '\n')

	// Read headers
	var l int
	for {
		if s, err = tp.ReadLine(); err != nil {
			return nil, nil, err
		}

		data = append(data, []byte(s)...)
		data = append(data, '\n')

		mime := strings.Split(s, ":")

		if len(mime) > 1 {

			if mime[1][0] == ' ' { mime[1] = mime[1][1:] }

			req.Header.Add(mime[0], mime[1])

			tmp := req.Header.Get("Content-Length")

			if len(tmp) > 0 {
				l, err = strconv.Atoi(tmp)
				if err != nil {
					return nil, nil, errors.New(
						fmt.Sprintf("malformed HTTP %s %s", s))
				}
			}

		} else {
			s = mime[0]
			break
		}
	}

	// Read body
	buf := make([]byte, l)

	if _, err := b.Read(buf); err != nil {
		return nil, nil, err
	}

	data = append(data, buf...)

	req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(data[:])))
	if err != nil {
		return nil, nil, err
	}

	return req, data, nil

}
