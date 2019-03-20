package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

const version = 1

func main() {
	flag.Parse()
	arguments := flag.Args()

	if len(arguments) > 0 {
		i := 0
		for {
			i++
			fmt.Printf("%d %s\n", i, strings.Join(arguments, " "))
			time.Sleep(5 * time.Second)
		}
	} else {
		log.Fatalln("Gimme some arguments!")
	}
}
