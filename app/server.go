package main

import (
	"bytes"
	"flag"
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

type Response struct {
	Version string
	Status string
	Headers map[string]string
	Body []byte
}

var dir = flag.String("directory", ".", "Directory to serve files from")

func main() {
	flag.Parse()
	fmt.Printf("Directory: %s\n", *dir)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Printf("Listening on %s\n", l.Addr())
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

func parseRequest(conn net.Conn) Request {
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
	return req
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	
	req := parseRequest(conn)
	
	if req.Method == "GET" {
		res := handleGetRequest(req)
		writeResponse(conn, res)
	} else if req.Method == "POST" {
		res := handlePostRequest(req)
		writeResponse(conn, res)
	} else {
		res := Response{ Version: "HTTP/1.1", Status: "405 Method Not Allowed"}
		writeResponse(conn, res)
	}
	
}

func handleGetRequest(req Request) Response {
	if req.Path == "/" {
		return Response{ Version: "HTTP/1.1", Status: "200 OK"}
	} else if strings.HasPrefix(req.Path, "/files/") {
		file := strings.TrimPrefix(req.Path, "/files/")
		fp := fmt.Sprint(*dir, string(os.PathSeparator), file)
		f, err := os.ReadFile(fp)
		if err != nil {
			return Response{ Version: "HTTP/1.1", Status: "404 Not Found"}
		}
		return Response{ Version: "HTTP/1.1", Status: "200 OK", Headers: map[string]string{"Content-Length": fmt.Sprintf("%d", len(f)), "Content-Type": "application/octet-stream"}, Body: f}
		
	} else if strings.HasPrefix(req.Path, "/echo/"){
		body := strings.TrimPrefix(req.Path, "/echo/")
		return Response{ Version: "HTTP/1.1", Status: "200 OK", Headers: map[string]string{"Content-Length": fmt.Sprintf("%d", len(body)), "Content-Type": "text/plain"}, Body: []byte(body)}
	} else if strings.HasPrefix(req.Path, "/user-agent") {
		ua := req.Headers["User-Agent"]
		return Response{ Version: "HTTP/1.1", Status: "200 OK", Headers: map[string]string{"Content-Length": fmt.Sprintf("%d", len(ua)), "Content-Type": "text/plain"}, Body: []byte(ua)}
	} else {
		return Response{ Version: "HTTP/1.1", Status: "404 Not Found"}
	}
}

func handlePostRequest(req Request) Response {
	if strings.HasPrefix(req.Path, "/files/") {
		fname := strings.TrimPrefix(req.Path, "/files/")
		fp := fmt.Sprint(*dir, string(os.PathSeparator), fname)
		err := os.WriteFile(fp, req.Body, 0644)
		if err != nil {
			return Response{ Version: "HTTP/1.1", Status: "500 Internal Server Error"}
		}
	}
	
	return Response{ Version: "HTTP/1.1", Status: "201 Created"}
}

func writeResponse(conn net.Conn, res Response) {
	_, err := conn.Write([]byte(res.Version + " " + res.Status + "\r\n"))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
	for k, v := range res.Headers {
		_, err := conn.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	}
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
	_, err = conn.Write(res.Body)
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
