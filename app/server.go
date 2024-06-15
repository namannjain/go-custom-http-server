package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// var statusCodes = map[int]string{
// 	200: "OK",
// 	404: "Not Found",
// }

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

// func handleError(err error, errorMsg string, osExitCode int) {
// 	if err != nil {
// 		fmt.Println(errorMsg, ": ", err.Error())
// 		os.Exit(osExitCode)
// 	}
// }

// func (r Response) createResponseString() string {
// 	statusText, ok := statusCodes[r.StatusCode]
// 	if !ok {
// 		statusText = "Unknown"
// 	}

// 	// No headers so assume plain text result.
// 	if r.Headers == nil {
// 		r.Headers = map[string]string{
// 			"Content-Type": "text/plain",
// 		}
// 	}

// 	// Figure out content length if not set.
// 	if _, ok = r.Headers["Content-Length"]; !ok {
// 		r.Headers["Content-Length"] = strconv.Itoa(len(r.Body))
// 	}
// 	var headerString strings.Builder
// 	for k, v := range r.Headers {
// 		headerString.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
// 	}
// 	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", r.StatusCode, statusText, headerString.String(), r.Body)
// }

// func parseRequest(req string) Request {
// 	request := Request{
// 		Headers: make(map[string]string),
// 	}
// 	methodPathAndHeaders := strings.Split(strings.Split(req, "\r\n\r\n")[0], "\r\n")
// 	methodAndPath := strings.Split(methodPathAndHeaders[0], " ")
// 	request.Method = methodAndPath[0]
// 	request.Path = methodAndPath[1]
// 	request.Headers = extractHeadersMap(methodPathAndHeaders[1:])
// 	request.Body = ""

// 	return request
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
	defer conn.Close()

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

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
