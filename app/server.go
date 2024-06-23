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

		answer(requestLine, connection)
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
func answer(requestLine string, conn net.Conn) {
	requestTarget := strings.Split(requestLine, " ")[1]
	userAgentHeader := strings.Split(requestLine, "\r\n")[3]
	fmt.Println("UserAgent String: ", userAgentHeader)
	if requestTarget == "/" {
		fmt.Println("Known target. Returning 200")
		// Send a HTTP 200 Response with HTTP/1.1
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(requestTarget, "/echo/") {
		fmt.Println("ECHO target. ")
		value := requestTarget[6:]
		fmt.Println("Value: ", value)
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(value)) + "\r\n\r\n" + value))
	} else if requestTarget == "/user-agent" {
		fmt.Println("User-Agent target. ")
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
