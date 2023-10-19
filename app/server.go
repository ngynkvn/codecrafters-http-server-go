package main

import (
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

func handleConn(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 32768)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}
	requestStr := string(buffer[:n])
	lines := strings.Split(requestStr, "\r\n")
	startLine := NewStartLine(lines[0])

	headerMap := map[string]string{}
	headerLines := lines[1:]
	for _, line := range headerLines {
		kv := strings.Split(line, ": ")
		if len(kv) == 2 {
			headerMap[kv[0]] = kv[1]
		}
	}

	switch true {
	case strings.HasPrefix(startLine.Path, "/echo/"):
		echoWord := strings.TrimPrefix(startLine.Path, "/echo/")
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(conn, "Content-Type: text/plain\r\n")
		fmt.Fprintf(conn, "Content-Length: %d\r\n", len(echoWord))
		fmt.Fprintf(conn, "\r\n")
		fmt.Fprintf(conn, "%s", echoWord)
		return
	case strings.HasPrefix(startLine.Path, "/user-agent"):
		resp := http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Content-Type": {"text/plain"},
			},
			ContentLength: int64(len(headerMap["User-Agent"])),
			Body:          io.NopCloser(strings.NewReader(headerMap["User-Agent"])),
		}
		resp.Write(conn)

	case startLine.Path == "/":
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}
