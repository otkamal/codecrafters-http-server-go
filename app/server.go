package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const BufferSize int = 1024
const HttpResponseOk string = "HTTP/1.1 200 OK\r\n\r\n"
const HttpResponseNotFound string = "HTTP/1.1 404 Not Found\r\n\r\n"

type HttpRequest struct {
	Method string
	Path   string
}

type HttpResponse struct {
	StatusLine string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, BufferSize)

	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error received:", err)
		return
	}

	fmt.Printf("Received: %v", string(buffer[:]))
	httpRequest := HttpRequestFromString(string(buffer[:]))

	_, err = conn.Write([]byte(CraftHttpResponse(httpRequest).StatusLine))
	if err != nil {
		fmt.Println("Error Sending:", err)
		return
	}
}

func HttpRequestFromString(request string) HttpRequest {

	tokenizedRequest := strings.Split(request, "\r\n")
	startLine := tokenizedRequest[0]

	// fmt.Println(startLine)
	// for _, token := range strings.Split(startLine, " ") {
	// 	fmt.Printf("%v\n", token)
	// }

	tokenizedStartLine := strings.Split(startLine, " ")

	return HttpRequest{Method: tokenizedStartLine[0], Path: tokenizedStartLine[1]}

}

func CraftHttpResponse(request HttpRequest) HttpResponse {
	var response HttpResponse
	switch request.Method {
	case "GET":
		switch request.Path {
		case "/":
			response.StatusLine = HttpResponseOk
		default:
			response.StatusLine = HttpResponseNotFound
		}
	}
	return response
}
