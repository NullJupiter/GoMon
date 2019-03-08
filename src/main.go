package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Get working directory flag
	workingDir := flag.String("p", "", "Provide the full path to your working directory! The program will monitor changes on all files but just restart a file named \"main.go\"!")
	flag.Parse()

	// Check if parameter was empty
	if *workingDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Start main.go for the first time
	restartingCmd := exec.Command("go", "run", fmt.Sprintf("%v/main.go", *workingDir))
	output, err := restartingCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Something went wrong!")
	}
	fmt.Printf("%s\n", output)

	// Monitor file changes
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer fileWatcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-fileWatcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// Print which file has changed
					log.Println("Modified file:", event.Name)
					log.Println("Restarting main.go ...")

					// Starting main.go after changes
					restartingCmd := exec.Command("go", "run", fmt.Sprintf("%v/main.go", *workingDir))
					output, err := restartingCmd.CombinedOutput()
					if err != nil {
						log.Fatal("Something went wrong!")
					}

					fmt.Printf("%s\n", output)
				}

			case err, ok := <-fileWatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = fileWatcher.Add(*workingDir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
