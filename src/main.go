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

func main() {
	running := true
	config := getConfig()
	restartingCmd := startCommand(config)

	go killOnSignal(&running, &restartingCmd)

	// Monitor file changes
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer fileWatcher.Close()

	go watchFiles(&running, &fileWatcher, &restartingCmd)

	for _, dir := range config.WatchDirectories {
		// log.Printf("Adding %v to fileWatcher\n", *workingDir)
		err = fileWatcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Wait forever
	done := make(chan bool)
	<-done
}

type Config struct {
	WatchDirectories []string
	Command          string
	CommandArguments []string
}

func watchFiles(running *bool, watcher *fsnotify.Watcher, restartingCmd **exec.Cmd) {
	cmd := *restartingCmd

	for running {
		select {
		case event, ok := <-fileWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// log.Println("Modified file:", event.Name)
				// log.Println("Restarting main.go ...")

				fmt.Println("Killing process ", restartingCmd.Process.Pid)
				syscall.Kill(-restartingCmd.Process.Pid, syscall.SIGKILL)
				restartingCmd.Wait()

				restartingCmd = startCommand(config)
			}

		case err, ok := <-fileWatcher.Errors:
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

	// Get working directory flag
	command := flag.String("cmd", "", "Provide a command to execute and restart. If nothing is set, this defaults to \"go run $path/*.go\"")
	workingDir := flag.String("p", "", "Provide the full path to your working directory")
	recursive := flag.Bool("r", true, "Search through the working directory recursively for file changes")
	flag.Parse()

	// Check if parameter was empty
	if *workingDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *command == "" {
		files, err := filepath.Glob(filepath.Join(*workingDir, "*.go"))
		if err != nil {
			log.Fatalf("Error looking fo go files: %s\n", err.Error())
		}

		if len(files) < 1 {
			log.Fatalf("Error: No go files found")
		}

		*command = "go run " + strings.Join(files, " ")
	}

	// TODO: What about paths with " " in their name?
	commandParts := strings.Split(*command, " ")
	// fmt.Printf("CMD %#v\n", commandParts)
	config.Command = commandParts[0]
	config.CommandArguments = commandParts[1:]

	config.WatchDirectories = make([]string, 0, 10)
	if *recursive {
		// Recursively search for folders
		// fmt.Print("Recursively searching ", *workingDir, "\n")
		findDirectories(*workingDir, &config.WatchDirectories)
	} else {
		config.WatchDirectories = append(config.WatchDirectories, *workingDir)
	}

	return config
}

func startCommand(config *Config) *exec.Cmd {
	restartingCmd := exec.Command(config.Command, config.CommandArguments...)
	restartingCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	restartingCmd.Stdout = os.Stdout

	// log.Printf("Starting %v/main.go\n", *workingDir)
	err := restartingCmd.Start()
	if err != nil {
		log.Fatal("Something went wrong!")
	}

	fmt.Println("Starting process ", restartingCmd.Process.Pid)

	return restartingCmd
}

func killOnSignal(running *bool, restartingCmd **exec.Cmd) {
	chanSigInt := make(chan os.Signal)
	signal.Notify(chanSigInt, syscall.SIGINT, syscall.SIGTERM)

	<-chanSigInt
	*running = false

	cmd := *restartingCmd

	fmt.Println("Signal caught, Killing process ", cmd.Process.Pid)
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	cmd.Wait()

	os.Exit(0)
}
