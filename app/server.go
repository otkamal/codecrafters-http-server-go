package main

import (
	"fmt"
	"net"
	"os"
)

const BufferSize int = 1024

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

	ok_response := []byte("HTTP/1.1 200 OK\r\n\r\n")
	_, err = conn.Write(ok_response)
	if err != nil {
		fmt.Println("Error Sending:", err)
		return
	}
}
