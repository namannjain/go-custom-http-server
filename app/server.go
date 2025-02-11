package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

var statusCodes = map[int]string{
	200: "OK",
	404: "Not Found",
	201: "Created",
}

var requestTypes = map[string]string{
	"GET":  "GET",
	"POST": "POST",
}

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func handleError(err error, errorMsg string, osExitCode int) {
	fmt.Println(errorMsg, err.Error())
	if osExitCode != -1 {
		os.Exit(osExitCode)
	}
}

func (res Response) createResponseString() string {
	statusText, ok := statusCodes[res.StatusCode]
	if !ok {
		statusText = "Unknown"
	}

	var headerString strings.Builder
	for k, v := range res.Headers {
		headerString.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", res.StatusCode, statusText, headerString.String(), res.Body)
}

func parseRequest(req string) Request {
	request := Request{
		Headers: make(map[string]string),
	}
	methodPathAndHeaders := strings.Split(strings.Split(req, "\r\n\r\n")[0], "\r\n")
	methodAndPath := strings.Split(methodPathAndHeaders[0], " ")
	request.Method = methodAndPath[0]
	request.Path = methodAndPath[1]
	request.Headers = extractHeadersMap(methodPathAndHeaders[1:])
	request.Body = []byte(strings.Trim(strings.Split(req, "\r\n\r\n")[1], "\x00"))

	return request
}

func extractHeadersMap(headers []string) map[string]string {
	headersMap := make(map[string]string)
	for _, headerStr := range headers {
		tokens := strings.Split(headerStr, ":")
		headersMap[strings.Trim(tokens[0], " ")] = strings.Trim(tokens[1], " ")
	}
	return headersMap
}

func CreateFileInDir(directory string, fileName string, fileData []byte) error {
	//assuming directory exists
	//create file
	file, err := os.Create(directory + fileName)
	if err != nil {
		return errors.New("error creating file")
	}
	defer file.Close()

	//write into file
	_, err = file.Write(fileData)
	if err != nil {
		return errors.New("error writing to file")
	}

	fmt.Println("Successfully created file with its content")
	return nil
}

func compressStringToGzip(data string) ([]byte, error) {
	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)
	defer gzipWriter.Close()

	_, err := gzipWriter.Write([]byte(data))
	if err != nil {
		var emptyArr []byte
		return emptyArr, fmt.Errorf("error writing data to Gzip writer: %w", err)
	}

	return gzipBuffer.Bytes(), nil
}

func decompressGzip(data []byte) []byte {
	reader := bytes.NewReader(data)
	gzipReader, _ := gzip.NewReader(reader)
	defer gzipReader.Close()

	decompressedData, _ := ioutil.ReadAll(gzipReader)
	return decompressedData
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	byteSize, _ := conn.Read(buffer)

	request := parseRequest(string(buffer[:byteSize]))
	response := Response{}

	switch {
	case request.Path == "/":
		response.StatusCode = 200
	case request.Path == "/user-agent":
		response.StatusCode = 200
		if userAgent, ok := request.Headers["User-Agent"]; ok {
			response.Body = userAgent
			response.Headers = map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(userAgent)),
			}
		}
	case strings.HasPrefix(request.Path, "/echo"):
		response.StatusCode = 200
		message := strings.SplitN(request.Path, "/echo/", 2)[1]
		response.Body = message
		response.Headers = map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(message)),
		}
	case strings.HasPrefix(request.Path, "/files") && request.Method == requestTypes["GET"]:
		response.StatusCode = 404
		fileName := strings.SplitN(request.Path, "/files/", 2)[1]
		dir := os.Args[2]
		fileData, err := os.ReadFile(dir + fileName)
		if err == nil {
			response.StatusCode = 200
			response.Body = string(fileData)
			response.Headers = map[string]string{
				"Content-Type":   "application/octet-stream",
				"Content-Length": strconv.Itoa(len(string(fileData))),
			}
		}
	case strings.HasPrefix(request.Path, "/files") && request.Method == requestTypes["POST"]:
		response.StatusCode = 404
		fileName := strings.SplitN(request.Path, "/files/", 2)[1]
		dir := os.Args[2]
		fileData := request.Body

		err := CreateFileInDir(dir, fileName, fileData)
		if err != nil {
			handleError(err, " ", -1)
			break
		} else {
			response.StatusCode = 201
		}
	default:
		response.StatusCode = 404
	}

	//handle encoding
	if acceptEncoding, ok := request.Headers["Accept-Encoding"]; ok && strings.Contains(acceptEncoding, "gzip") && response.StatusCode != 404 {
		response.Headers["Content-Encoding"] = "gzip"
		gzipBuffer, err := compressStringToGzip(response.Body)
		if err != nil {
			handleError(err, " ", -1)
		} else {
			fmt.Println(string(decompressGzip(gzipBuffer)))
			response.Body = string(gzipBuffer)
			response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))
		}
	}

	conn.Write([]byte(response.createResponseString()))
}

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		handleError(err, "Failed to bind to port 4221: ", 1)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			handleError(err, "Error accepting connection: ", 1)
		}

		go handleConnection(conn)
	}
}
