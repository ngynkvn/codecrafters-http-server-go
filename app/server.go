package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

type StartLine struct {
	Method   string
	Path     string
	Protocol string
}

func NewStartLine(str string) StartLine {
	words := strings.Split(str, " ")
	return StartLine{
		words[0],
		words[1],
		words[2],
	}
}

var directory string

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := bufio.NewReader(conn)
	request, err := http.ReadRequest(buf)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}

	switch true {
	case strings.HasPrefix(request.URL.Path, "/echo/"):
		echoWord := strings.TrimPrefix(request.URL.Path, "/echo/")
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(conn, "Content-Type: text/plain\r\n")
		fmt.Fprintf(conn, "Content-Length: %d\r\n", len(echoWord))
		fmt.Fprintf(conn, "\r\n")
		fmt.Fprintf(conn, "%s", echoWord)
		return
	case strings.HasPrefix(request.URL.Path, "/files/") && request.Method == http.MethodPost:
		filePath := strings.TrimPrefix(request.URL.Path, "/files/")
		writePath := fmt.Sprintf("%s/%s", directory, filePath)
		file, _ := os.Create(writePath)
		bytes, _ := io.ReadAll(request.Body)
		file.Write(bytes)
		fmt.Fprintf(conn, "HTTP/1.1 201 CREATED\r\n\r\n")
		return
	case strings.HasPrefix(request.URL.Path, "/files/") && request.Method == http.MethodGet || request.Method == "":
		filePath := strings.TrimPrefix(request.URL.Path, "/files/")
		writePath := fmt.Sprintf("%s/%s", directory, filePath)
		if file, err := os.Open(writePath); os.IsNotExist(err) {
			fmt.Fprintf(conn, "HTTP/1.1 404 NOT FOUND\r\n\r\n")
		} else {
			bytes, _ := io.ReadAll(file)
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
			fmt.Fprintf(conn, "Content-Type: application/octet-stream\r\n")
			fmt.Fprintf(conn, "Content-Length: %d\r\n", len(bytes))
			fmt.Fprintf(conn, "\r\n")
			fmt.Fprintf(conn, "%s", bytes)
		}
		return
	case strings.HasPrefix(request.URL.Path, "/user-agent"):
		agent := request.Header["User-Agent"][0]
		resp := http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Content-Type": {"text/plain"},
			},
			ContentLength: int64(len(agent)),
			Body:          io.NopCloser(strings.NewReader(agent)),
		}
		resp.Write(conn)

	case request.URL.Path == "/":
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
		return
	default:
		fmt.Fprintf(conn, "HTTP/1.1 404 NOT FOUND\r\n\r\n")
		return
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	flag.StringVar(&directory, "directory", "", "directory")
	flag.Parse()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}
