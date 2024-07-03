package main

import (
	"errors"
	"fmt"
	"strconv"
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
	method, requestTarget := strings.Split(requestLine, " ")[0], strings.Split(requestLine, " ")[1]
	// Save headers in map
	headers := make(map[string]string)
	for _, header := range strings.Split(requestLine, "\r\n") {
		if strings.Contains(header, ":") {
			headerParts := strings.Split(header, ":")
			headers[headerParts[0]] = headerParts[1]
		}
	}
	requestBody := strings.Split(requestLine, "\r\n\r\n")[1]
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
		fmt.Println("Directory: ", directory)
		fmt.Print("Method", method)
		if method == "GET" {
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
		} else if method == "POST" {
			// if strings.Contains(headers["Content-Type"], "application/octet-stream") {
			// 	fmt.Println("Invalid Content-Type. Returning 400")
			// 	fmt.Println("Content-Type: ", headers["Content-Type"])
			// 	conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			// 	conn.Close()
			// 	return
			// }
			//print all headers
			for key, value := range headers {
				fmt.Println("Key: ", key, "Value: ", value)
			}
			// Save content length from the headers in a variable. If it is not present, return 400. Content-Length is mandatory and is a number
			contentLength := headers["Content-Length"]
			println("Content-Length: ", contentLength)
			if contentLength == "" {
				fmt.Println("Invalid Content-Length. Returning 400")
				fmt.Println("Content-Length: ", headers["Content-Length"])
				conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
				conn.Close()
				return
			}
			fileSize, err := strconv.ParseInt(strings.Trim(contentLength, " "), 10, 64)
			fmt.Println("File Size: ", fileSize)
			if err != nil {
				fmt.Println("Error in converting content-length to filesize")
				fmt.Println("Error", err.Error())
				conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
				return
			}
			fileName := directory + requestTarget[7:]
			fmt.Println("File Name: ", fileName)
			// Create File. By default, it will overwrite the file if it exists
			// There are fileSize bytes to be read from the request body
			file, err := os.Create(fileName)
			if err != nil {
				fmt.Println("Error creating file: ", err.Error())
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				conn.Close()
				return
			}
			fmt.Println("File created successfully")
			// Write the file
			_, err = file.Write([]byte(requestBody))
			fmt.Println("File content: ", string(requestBody))
			if err != nil {
				fmt.Println("Error writing file: ", err.Error())
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				conn.Close()
				return
			}
			fmt.Println("File wrote successfully. 201 Created")
			// Close the file
			file.Close()
			fmt.Println("File closed successfully")
			// Send a HTTP 201 Response with HTTP/1.1
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			conn.Close()

		}
	} else if requestTarget == "/user-agent" {
		fmt.Println("User-Agent target. ")
		fmt.Println("User-Agent Header: ", headers["User-Agent"])
		if headers["User-Agent"] == "" {
			fmt.Println("User-Agent header not present. Returning 400")
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		}
		userAgent := headers["User-Agent"]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
	} else {
		fmt.Println("Unknown target. Returning 404")
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
