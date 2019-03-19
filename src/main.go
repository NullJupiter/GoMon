package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

// Config is a type to configure the program.
type Config struct {
	WatchDirectories []string
	Command          string
	CommandArguments []string
}

// Global variables.
var running = true
var config = getConfig()
var restartingCmd = startCommand()

func main() {

	go killOnSignal()

	// Monitor file changes
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer fileWatcher.Close()

	go watchFiles(fileWatcher)

	for _, dir := range config.WatchDirectories {
		err = fileWatcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Wait forever
	done := make(chan bool)
	<-done
}

func watchFiles(watcher *fsnotify.Watcher) {
	for running {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Modified file:", event.Name)
				log.Print("Restarting go program ...\n\n")

				syscall.Kill(-restartingCmd.Process.Pid, syscall.SIGKILL)
				restartingCmd.Wait()

				restartingCmd = startCommand()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func findDirectories(dir string, directories *[]string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("Error reading directory %s: %s\n", dir, err.Error())
	}

	*directories = append(*directories, dir)
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() && fileName[0] != '.' {
			findDirectories(filepath.Join(dir, fileName), directories)
		}
	}
}

func getConfig() *Config {
	config := &Config{}

	// Get flags
	command := flag.String("cmd", "", "Provide a command to execute and restart. If nothing is set, this defaults to \"go run $path/*.go\"")
	workingDir := flag.String("p", "", "Provide the full path to your working directory")
	recursive := flag.Bool("r", true, "Search through the working directory recursively for file changes (set to true or false")
	flag.Parse()

	// Check if parameters are empty
	if *workingDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *command == "" {
		files, err := filepath.Glob(filepath.Join(*workingDir, "*.go"))
		if err != nil {
			log.Fatalf("Error looking for go files: %s\n", err.Error())
		}

		if len(files) < 1 {
			log.Fatalf("Error: No go files found")
		}

		*command = "go run " + strings.Join(files, " ")
	}

	// TODO: What about paths with " " in their name?
	commandParts := strings.Split(*command, " ")
	config.Command = commandParts[0]
	config.CommandArguments = commandParts[1:]

	config.WatchDirectories = make([]string, 0, 10)
	if *recursive {
		// Recursively search for folders
		findDirectories(*workingDir, &config.WatchDirectories)
	} else {
		config.WatchDirectories = append(config.WatchDirectories, *workingDir)
	}

	return config
}

func startCommand() *exec.Cmd {
	restartingCmd := exec.Command(config.Command, config.CommandArguments...)
	restartingCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	restartingCmd.Stdout = os.Stdout

	err := restartingCmd.Start()
	if err != nil {
		log.Fatal("Something went wrong!")
	}

	return restartingCmd
}

func killOnSignal() {
	chanSigInt := make(chan os.Signal)
	signal.Notify(chanSigInt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-chanSigInt
	running = false

	fmt.Println("Signal caught, Killing process", restartingCmd.Process.Pid)

	err := syscall.Kill(-restartingCmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		killOnSignal()
	}

	restartingCmd.Wait()

	os.Exit(0)
}
