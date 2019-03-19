package main

import (
	"fmt"
	"time"
)

func output() {
	time.Sleep(time.Second)
	fmt.Print("!")
	for {
		time.Sleep(time.Second)
		fmt.Print(".")
	}
}
