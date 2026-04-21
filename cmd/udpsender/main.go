package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	updAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving UDP address: %v\n", err)
		os.Exit(1)
	}

	// to prepare a UDP connection, and defer the closing of the connection.
	conn, err := net.DialUDP("udp", nil, updAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing UDP: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", "localhost:42069")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending message: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Message sent: %s", message)
	}
}
