package main

import (
	"os"
	"fmt"
	"flag"
	"bytes"
	"net"
	"time"
	"strconv"
	"errors"
	"bufio"
	"strings"
	"crypto/tls"
//	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/http/httputil"
	"compress/gzip"
)

const (
	MODE_CONN_TIMEOUT = 0
	MODE_HTTP_500     = 1
	MODE_HTTP_PROXY   = 2
)

var (
	domain   = flag.String ("domain", "example.com:443", "")
	listen   = flag.String ("listen", ":8000",           "")
	encode   = flag.String ("encode",  "gzip",           "only support [gzip|none|\"\"]")
	mode     = flag.Int    ("mode",         0,           "0 [conn timeout] | 1 [http 500] | 2 [proxy]")
	withTLS  = flag.Bool   ("tls",      false,           "With TLS")
	insecure = flag.Bool   ("k",        false,           "TLS Insecure Skip Verify")
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
	req, dat, err := readRequest(bufio.NewReader(conn))
	if err != nil {
		conn.Close()
		fmt.Println(err)
		return
	}

	fmt.Printf("-\n%s-\n", dat)

	dump, _ := httputil.DumpRequest(req, false)

	fmt.Printf("\n-\n%s-\n", dump)

	var resp []byte

	switch m := *mode; m {
	case MODE_HTTP_500:
		response := http.Response{
					StatusCode:    500,
					ProtoMajor:    req.ProtoMajor,
					ProtoMinor:    req.ProtoMinor,
					Request:       req,
					Header:        http.Header{},
//					Body:          ioutil.NopCloser(bytes.NewReader(dat[:])),
					ContentLength: -1,
		}

		var braw bytes.Buffer
		response.Write(&braw)

		resp = braw.Bytes()

		fmt.Println("  ---  ")
		fmt.Println(string(resp))

		conn.Write(resp[:])
		conn.Close()


	case MODE_HTTP_PROXY:
		conf := &tls.Config{
				InsecureSkipVerify: *insecure,
		}

		var sock net.Conn
		var err  error

		if *withTLS {
			sock, err = tls.Dial("tcp", *domain, conf)
		} else {
			sock, err = net.Dial("tcp", *domain)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		sock.Write(dat[:])

		chunk := make([]byte, 100)

		var buf []byte

		fmt.Println("  - INIT -  ")
		for {
			sock.SetDeadline(time.Now().Add(5 * time.Second))
			n, err := sock.Read(chunk)

			conn.Write(chunk[:n])
			buf = append(buf, chunk[:n]...)

			if err != nil { break }
		}

		fmt.Printf("%s\n", buf)

		enc, _ := getEncondeType(buf)

		fmt.Printf("  - DECODE DATA %s -  \n", enc)

		if enc == "gzip" {

			data, _ := gUnzipData(buf)
			fmt.Println(string(data))

		}

		fmt.Println("  - END -  ")

		sock.Close()
		conn.Close()


	default:
		// nothing
	}

}

func getEncondeType(data []byte) (encode string, err error) {

	tp  := textproto.NewReader(bufio.NewReader(bytes.NewReader(data)))

	// remove first line
	if _, err = tp.ReadLine(); err != nil {
		return "", err
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return "", err
	}

	return mimeHeader.Get("Content-Encoding"), nil

}

func gUnzipData(data []byte) (resData []byte, err error) {


	var _data bytes.Buffer
	buf := string(data)


	r := strings.SplitN(buf, "\r\n\r\n", 3)

	// remove len
	if len(r) > 1 {
		r = strings.SplitN(r[1], "\r\n", 3)
	}

	// no report error, FIX THIS .
	if len(r) > 1 {

		b := bytes.NewBuffer([]byte(r[1]))

		reader, err := gzip.NewReader(b)
		if err != nil {
			return nil, err
		}


		if _, err := _data.ReadFrom(reader); err != nil {
			return nil, err
		}

	}


	return _data.Bytes(), nil
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
	data = append(data, []byte("\r\n")...)

	// Read headers
	var l int
	for {
		if s, err = tp.ReadLine(); err != nil {
			return nil, nil, err
		}


		mime := strings.Split(s,       ":")
		addr := strings.Split(*domain, ":")

		if len(mime) > 1 {

			if mime[0] == "Upgrade-Insecure-Requests" { continue }

			if mime[0] == "Accept-Encoding" && *encode != "" {
				fmt.Println("NO SUPPORT ENCODE", s)
				s = fmt.Sprintf("%s: %s", mime[0], *encode)
			}

			if mime[0] == "Host" {
				s = fmt.Sprintf("%s: %s", mime[0], addr[0])
			}
		}

		data = append(data, []byte(s)...)
		data = append(data, []byte("\r\n")...)

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


		} else { break }
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
