package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type ResponseTemplate struct {
	Body    string
	Headers map[string]string
}

// func createResponseString(body string) string {
// 	responseTemplateStr := `HTTP/1.1 200 OK\r\n{{ range $key, $value := .Headers }}{{$key}}: {{$value}}\r\n{{ end }}\r\n{{ .Body }}`
// 	tmpl := template.Must(template.New("example").Parse(responseTemplateStr))

// 	responseTemplate := ResponseTemplate{
// 		Body: body,
// 		Headers: map[string]string{
// 			"Content-Type":   "text/plain",
// 			"Content-Length": strconv.FormatInt(int64(len(body)), 10),
// 		},
// 	}

// 	var buff bytes.Buffer

// 	if err := tmpl.Execute(&buff, responseTemplate); err != nil {
// 		fmt.Println("Error executing text template: ", err)
// 		os.Exit(2)
// 	}

// 	return buff.String()
// }

func extractHeadersMap(headers []string) map[string]string {
	headersMap := make(map[string]string)
	for _, headerStr := range headers {
		tokens := strings.Split(headerStr, ":")
		headersMap[strings.Trim(tokens[0], " ")] = strings.Trim(tokens[1], " ")
	}
	return headersMap
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	byteSize, _ := conn.Read(buffer)
	request := string(buffer[:byteSize])

	requestAndHeaders := strings.Split(request, "\r\n\r\n")
	requestAndHeaders = strings.Split(requestAndHeaders[0], "\r\n")
	requestLine := requestAndHeaders[0]
	headersLine := requestAndHeaders[1:]
	path := strings.Split(requestLine, " ")[1]
	splitPath := strings.Split(path, "/")

	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if splitPath[1] == "echo" {
		message := splitPath[2]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else if splitPath[1] == "user-agent" {
		headersMap := extractHeadersMap(headersLine)
		if val, ok := headersMap["User-Agent"]; ok {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)))
		}
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
