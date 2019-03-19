package main

import (
	"fmt"
	"time"
)

const version = 1

var delay = 5 * time.Second

func main() {
	fmt.Printf("Test started. Exiting in %d seconds...\n", delay/1000000000)
	time.Sleep(delay)
	fmt.Println("Test exited")
}
