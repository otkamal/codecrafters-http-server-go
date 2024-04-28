package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const BufferSize int = 1024

const HttpResponseOk string = "HTTP/1.1 200 OK\r\n"
const HttpResponseNotFound string = "HTTP/1.1 404 Not Found\r\n\r\n"
const PlainTextResponse string = "Content-Type: text/plain\r\n"
const ContentLengthResponse string = "Content-Length: "

type HttpRequest struct {
	Method    string
	Path      string
	UserAgent string
}

type HttpResponse struct {
	StatusLine      string
	ResponseHeaders string
	ResponseBody    string
	BodyLength      int
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

	//fmt.Printf("Received: %v", string(buffer[:]))
	httpRequest := HttpRequestFromString(string(buffer[:]))
	test := []byte(CraftHttpResponse(httpRequest))
	fmt.Printf("====== STREAMED ======\n%c\n\n%v\n", test, test)
	_, err = conn.Write(test)
	if err != nil {
		fmt.Println("Error Sending:", err)
		return
	}
}

func HttpRequestFromString(request string) HttpRequest {

	tokenizedRequest := strings.Split(request, "\r\n")
	startLine := tokenizedRequest[0]

	var userAgent string
	userAgent = tokenizedRequest[2]
	userAgent = strings.Split(userAgent, ":")[1]
	userAgent = strings.Trim(userAgent, " ")

	tokenizedStartLine := strings.Split(startLine, " ")

	return HttpRequest{Method: tokenizedStartLine[0], Path: tokenizedStartLine[1], UserAgent: userAgent}

}

func CraftHttpResponse(request HttpRequest) string {
	var response HttpResponse
	var httpResponse string
	switch request.Method {
	case "GET":

		if request.Path == "/" {

			response.StatusLine = HttpResponseOk
			httpResponse = response.StatusLine + "\r\n"

		} else {

			//parse the path
			tokenizedPath := strings.Split(request.Path, "/")[1:]

			if tokenizedPath[0] == "echo" {
				response.StatusLine = HttpResponseOk
				response.ResponseBody = strings.Join(tokenizedPath[1:], "/")
				response.BodyLength = len(response.ResponseBody)

				httpResponse = response.StatusLine +
					PlainTextResponse +
					ContentLengthResponse + fmt.Sprintf("%v", response.BodyLength) + "\r\n\r\n" +
					response.ResponseBody

				// if response.ResponseBody != "" {
				// 	httpResponse += "\r\n"
				// }

				fmt.Printf("====== RESPONSE ======\n%v\n", httpResponse)

			} else if tokenizedPath[0] == "user-agent" {
				response.StatusLine = HttpResponseOk
				response.ResponseBody = request.UserAgent
				response.BodyLength = len(request.UserAgent)

				httpResponse = response.StatusLine +
					PlainTextResponse +
					ContentLengthResponse + fmt.Sprintf("%v", response.BodyLength) + "\r\n\r\n" +
					response.ResponseBody

			} else {
				response.StatusLine = HttpResponseNotFound
				httpResponse = response.StatusLine
			}

		}

	}

	return httpResponse
}
