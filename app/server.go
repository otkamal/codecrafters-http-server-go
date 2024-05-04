package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const BufferSize int = 1024

const HttpResponseOk string = "HTTP/1.1 200 OK\r\n"
const HttpResponseCreated string = "HTTP/1.1 201 Created\r\n\r\n"
const HttpResponseNotFound string = "HTTP/1.1 404 Not Found\r\n\r\n"

const PlainTextResponse string = "Content-Type: text/plain\r\n"
const ApplicationResponse string = "Content-Type: application/octet-stream\r\n"

const ContentLengthResponse string = "Content-Length: "

type HttpRequest struct {
	Method    string
	Path      string
	UserAgent string
	Body      string
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

	var dir = flag.String("directory", "", "Directory for file")
	flag.Parse()

	fmt.Printf("Home Directory: %v\n", *dir)

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

		go handleClient(conn, *dir)

	}

}

func handleClient(conn net.Conn, baseDirectory string) {

	defer conn.Close()

	buffer := make([]byte, BufferSize)

	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error received:", err)
		return
	}

	fmt.Printf("====== RECEIVED ======\n %v\n\n", buffer[:])

	httpRequest := HttpRequestFromString(strings.TrimSuffix(string(buffer[:]), "\r\n"))
	test := CraftHttpResponse(httpRequest, baseDirectory)

	fmt.Printf("\n====== STREAMED ======\n%c\n\n%v\n", test, test)

	_, err = conn.Write(test)
	if err != nil {
		fmt.Println("Error Sending:", err)
		return
	}

}

func HttpRequestFromString(request string) HttpRequest {

	request = strings.TrimSuffix(request, "\r\n")

	tokenizedRequest := strings.Split(request, "\r\n")
	startLine := tokenizedRequest[0]

	// DEBUG MESSAGES
	fmt.Printf("====== REQUEST ======\n")
	for i, val := range tokenizedRequest {
		fmt.Printf("%v: %v", i, val)
		if val == "" {
			fmt.Printf("EMPTY")
		}
		fmt.Printf("\n")
	}

	var userAgent string
	var body string
	// parsing is messed up and need to start from a later index in a situation where
	// we are not given headers...
	// this really needs to be fixed
	if len(tokenizedRequest) > 4 {
		userAgent = tokenizedRequest[2]
		userAgent = strings.Split(userAgent, ":")[1]
		userAgent = strings.Trim(userAgent, " ")
		body = tokenizedRequest[6]
	}

	tokenizedStartLine := strings.Split(startLine, " ")

	// if tokenizedStartLine[0] == "POST" {

	// }

	return HttpRequest{Method: tokenizedStartLine[0], Path: tokenizedStartLine[1], UserAgent: userAgent, Body: body}

}

func CraftHttpResponse(request HttpRequest, baseDirectory string) []byte {

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

			} else if tokenizedPath[0] == "files" {

				file, err := os.ReadFile(baseDirectory + "/" + strings.Join(tokenizedPath[1:], "/"))

				if err != nil {

					fmt.Println("Error Reading File: ", err.Error())

					response.StatusLine = HttpResponseNotFound
					httpResponse = response.StatusLine

				} else {

					fmt.Printf("\nRequested File: %v\n", baseDirectory+"/"+strings.Join(tokenizedPath[1:], "/"))
					response.StatusLine = HttpResponseOk
					response.ResponseBody = string(file)
					fmt.Printf("====== FILE CONTENTS ======\n%v", response.ResponseBody)
					response.BodyLength = len(response.ResponseBody)

					httpResponse = response.StatusLine +
						ApplicationResponse +
						ContentLengthResponse + fmt.Sprintf("%v", response.BodyLength) + "\r\n\r\n" +
						response.ResponseBody

				}

			} else {
				response.StatusLine = HttpResponseNotFound
				httpResponse = response.StatusLine
			}

		}

	case "POST":

		tokenizedPath := strings.Split(request.Path, "/")[1:]
		request.Body = strings.Replace(request.Body, "\x00", "", -1)
		fmt.Printf("\n\n%v\n\n", request.Body)
		err := os.WriteFile(baseDirectory+"/"+strings.Join(tokenizedPath[1:], "/"), []byte(request.Body), 0644)

		if err != nil {
			fmt.Println("Issue writing to file: ", err)
		}

		response.StatusLine = HttpResponseCreated
		httpResponse = response.StatusLine

	}

	return []byte(httpResponse)
}
