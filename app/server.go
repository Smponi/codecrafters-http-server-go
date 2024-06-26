package main

import (
	"errors"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage

	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Read the arguments of the program
	arguments := os.Args[1:]
	// Directory can be specified with --directory followed by the folder. Example: --directory /tmp/
	// Search for --directory
	// If found, set the directory to the next argument
	directory := ""
	for i := 0; i < len(arguments); i++ {
		if arguments[i] == "--directory" {
			directory = arguments[i+1]
			break
		}
	}
	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "localhost:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	// Wrap this in a while loop to handle up to 10 requests
	for i := 0; i < 10; i++ {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// Log all information of connection
		fmt.Println(connection.RemoteAddr().String())

		// Log request target
		// Read the first line of the request
		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			fmt.Println("Error reading request: ", err.Error())
			os.Exit(1)
		}
		requestLine := string(buf[:n])
		// Split request line
		// The String will looke like: "HTTP-Method /path/to/resource HTTP-Version"
		// We need to split it by space and get the second element

		answer(requestLine, connection, directory)
	}
}

func extractUserAgent(userAgentString string) (string, error) {
	if !strings.HasPrefix(userAgentString, "User-Agent") {
		return "", errors.New("User-Agent header not found")
	}
	userAgent := userAgentString[12:]
	fmt.Println("User-Agent: ", userAgent)
	return userAgent, nil
}

// This function gets the request path
func answer(requestLine string, conn net.Conn, directory string) {
	requestTarget := strings.Split(requestLine, " ")[1]
	// Getting User-Agent Header
	userAgentHeader := ""
	for _, header := range strings.Split(requestLine, "\r\n") {
		if strings.HasPrefix(header, "User-Agent") {
			fmt.Println("User-Agent Header: ", header)
			userAgentHeader = header
			break
		}
	}
	if requestTarget == "/" {
		fmt.Println("Known target. Returning 200")
		// Send a HTTP 200 Response with HTTP/1.1
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(requestTarget, "/echo/") {
		fmt.Println("ECHO target. ")
		value := requestTarget[6:]
		fmt.Println("Value: ", value)
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(value)) + "\r\n\r\n" + value))
	} else if strings.HasPrefix(requestTarget, "/files/") {
		fmt.Println("FILES target. ")
		if directory == "" {
			fmt.Println("No directory specified. Returning 500")
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		fileName := directory + requestTarget[7:]
		// Check if file exists
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			fmt.Println("File not found. Returning 404")
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		// Read file
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println("Error opening file: ", err.Error())
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("Error getting file info: ", err.Error())
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		fileSize := fileInfo.Size()
		// Read the file
		fileContent := make([]byte, fileSize)
		_, err = file.Read(fileContent)
		if err != nil {
			fmt.Println("Error reading file: ", err.Error())
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		//Print the file content
		fmt.Println("File content: ", string(fileContent))
		// Close the file
		file.Close()

		// Send file. Content-Type is application/octet-stream
		// Content-Length is the size of the file in bytes
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + fmt.Sprint(fileSize) + "\r\n\r\n" + string(fileContent)))

	} else if requestTarget == "/user-agent" {
		fmt.Println("User-Agent target. ")
		fmt.Println("User-Agent Header: ", userAgentHeader)
		userAgent, err := extractUserAgent(userAgentHeader)
		if err != nil {
			fmt.Println("Error extracting User-Agent: ", err.Error())
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
	} else {
		fmt.Println("Unknown target. Returning 404")
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
