package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func GetLinesFromReader(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		data := make([]byte, 8)
		str := ""

		for {

			n, err := f.Read(data)
			if err != nil {
				break
			}
			data := data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}

			str += string(data)
		}

		if len(str) != 0 {
			out <- str
		}

	}()

	return out
}

func main() {
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close() 
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("error", err)
		}

		fmt.Println("A connection has been accepted")

		for line := range GetLinesFromReader(conn) {
			fmt.Printf("read: %s\n", line)
		}

		fmt.Println("Connection closed")
	}

}
