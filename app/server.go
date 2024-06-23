package main

import (
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
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

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
	requestTarget := strings.Split(requestLine, " ")[1]
	fmt.Println(requestTarget)
	if requestTarget == "/" {
		fmt.Println("Known target. Returning 200")
		// Send a HTTP 200 Response with HTTP/1.1
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		fmt.Println("Unknown target. Returning 404")
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

}
