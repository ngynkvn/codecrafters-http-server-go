package main

import (
	"fmt"
	"net"
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

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
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
	switch true {
	case strings.HasPrefix(startLine.Path, "/echo/"):
		echoWord := strings.TrimPrefix(startLine.Path, "/echo/")
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(conn, "Content-Type: text/plain\r\n")
		fmt.Fprintf(conn, "Content-Length: %d\r\n", len(echoWord))
		fmt.Fprintf(conn, "\r\n")
		fmt.Fprintf(conn, "%s\r\n\r\n", echoWord)
		return
	case startLine.Path == "/":
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	default:
		fmt.Fprintf(conn, "HTTP/1.1 404 NOT FOUND\r\n\r\n")
	}
}
