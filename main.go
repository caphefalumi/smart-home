package main

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	// Configure the serial port
	config := &serial.Config{
		Name:        "COM7",
		Baud:        9600,
		ReadTimeout: time.Second * 5,
	}

	// Open the serial port
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Error opening port:", err)
	}
	defer port.Close()
	fmt.Println("Port opened")

	// Create a scanner to read lines (like ReadlineParser in Node.js)
	scanner := bufio.NewScanner(port)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Parsed Data:", line)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from port:", err)
	}
}
