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
		requestTarget := strings.Split(requestLine, " ")[1]
		fmt.Println(requestTarget)
		answer(requestTarget, connection)
	}
}

// This function gets the request path
func answer(requestTarget string, conn net.Conn) {
	if requestTarget == "/" {
		fmt.Println("Known target. Returning 200")
		// Send a HTTP 200 Response with HTTP/1.1
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(requestTarget, "/echo/") {
		fmt.Println("ECHO target. ")
		value := requestTarget[6:]
		fmt.Println("Value: ", value)
		// Response should look like this
		// // Status line
		// HTTP/1.1 200 OK
		// \r\n                          // CRLF that marks the end of the status line

		// // Headers
		// Content-Type: text/plain\r\n  // Header that specifies the format of the response body
		// Content-Length: 3\r\n         // Header that specifies the size of the response body, in bytes
		// \r\n                          // CRLF that marks the end of the headers

		// // Response body
		// abc                           // The string from the request
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(value)) + "\r\n\r\n" + value))
	} else {
		fmt.Println("Unknown target. Returning 404")
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
