package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	RequestTarget string
	Headers       map[string]string
	Body          string
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	arguments := os.Args[1:]
	directory := ""
	for i := 0; i < len(arguments); i++ {
		if arguments[i] == "--directory" {
			directory = arguments[i+1]
			break
		}
	}

	StartServer(directory)
}

func StartServer(directory string) {
	l, err := net.Listen("tcp", "localhost:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(connection, directory)
	}
}

func handleConnection(conn net.Conn, directory string) {
	defer fmt.Println("Closing connection")

	fmt.Println(conn.RemoteAddr().String())

	request, err := parseRequest(conn)
	if err != nil {
		fmt.Println("Error reading request:", err.Error())
		return
	}

	answer(request, conn, directory)
}

func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed request line")
	}

	method, requestTarget := parts[0], parts[1]
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}
	}

	body := ""
	if lengthStr, ok := headers["Content-Length"]; ok {
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, err
		}
		bodyBytes := make([]byte, length)
		_, err = reader.Read(bodyBytes)
		if err != nil {
			return nil, err
		}
		body = string(bodyBytes)
	}

	return &Request{Method: method, RequestTarget: requestTarget, Headers: headers, Body: body}, nil
}

func answer(request *Request, conn net.Conn, directory string) {
	switch {
	case request.RequestTarget == "/":
		sendResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
	case strings.HasPrefix(request.RequestTarget, "/echo/"):
		value := request.RequestTarget[6:]
		sendResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(value), value))
	case strings.HasPrefix(request.RequestTarget, "/files/"):
		handleFileRequest(request, conn, directory)
	case request.RequestTarget == "/user-agent":
		userAgent := strings.TrimSpace(request.Headers["User-Agent"])
		if userAgent == "" {
			sendResponse(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
		} else {
			sendResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent))
		}
	default:
		sendResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}

func handleFileRequest(request *Request, conn net.Conn, directory string) {
	if directory == "" {
		sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		return
	}

	fileName := directory + request.RequestTarget[7:]
	if request.Method == "GET" {
		sendFile(conn, fileName)
	} else if request.Method == "POST" {
		saveFile(conn, fileName, request.Body)
	} else {
		sendResponse(conn, "HTTP/1.1 405 Method Not Allowed\r\n\r\n")
	}
}

func sendFile(conn net.Conn, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			sendResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
		} else {
			sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		return
	}

	fileSize := fileInfo.Size()
	fileContent := make([]byte, fileSize)
	_, err = file.Read(fileContent)
	if err != nil {
		sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		return
	}

	sendResponse(conn, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", fileSize, string(fileContent)))
}

func saveFile(conn net.Conn, fileName, content string) {
	file, err := os.Create(fileName)
	if err != nil {
		sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	if err != nil {
		sendResponse(conn, "HTTP/1.1 500 Internal Server Error\r\n\r\n")
		return
	}

	sendResponse(conn, "HTTP/1.1 201 Created\r\n\r\n")
}

func sendResponse(conn net.Conn, response string) {
	fmt.Println(response)
	conn.Write([]byte(response))
	ererror := conn.Close()
	if ererror != nil {
		fmt.Println("Error closing connection:", ererror.Error())
	}
}
