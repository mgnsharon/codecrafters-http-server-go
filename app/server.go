package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method string
	Path string
	Version string
	Headers map[string]string
	Body []byte
}

func main() {
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Printf("Listening on %s", l.Addr())
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	b, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		os.Exit(1)
	}

	data := bytes.Split(buf[:b], []byte("\r\n\r\n"))
	header := bytes.Split(data[0], []byte("\r\n"))
	reqbody := data[1]
	sl := bytes.Split(header[0], []byte(" "))
	headers := header[1:]
	
	req := Request{
		Method: string(sl[0]),
		Path: string(sl[1]),
		Version: string(sl[2]),
		Body: reqbody,
	}

	req.Headers = make(map[string]string)
	for _, h := range headers {
		hh := bytes.Split(h, []byte(": "))
		req.Headers[string(hh[0])] = string(hh[1])
	}
	if req.Path == "/" {
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else if strings.HasPrefix(req.Path, "/echo/"){
		body := strings.TrimPrefix(req.Path, "/echo/")
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(body)) + "\r\n\r\n" + body))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else if strings.HasPrefix(req.Path, "/user-agent") {
		ua := req.Headers["User-Agent"] 
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(ua)) + "\r\n\r\n" + ua))
			if err != nil {
				fmt.Println("Error writing to connection: ", err.Error())
				os.Exit(1)
			}
	} else {
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	}
	
}
