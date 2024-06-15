package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	byteSize, _ := conn.Read(buffer)
	request := string(buffer[:byteSize])

	headers := strings.Split(request, "\r\n")
	path := strings.Split(headers[0], " ")[1]
	splitPath := strings.Split(path, "/")

	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if splitPath[1] == "echo" {
		message := splitPath[2]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else if splitPath[1] == "user-agent" {
		userAgent := strings.Split(headers[3], ": ")[1]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConnection(conn)
}
