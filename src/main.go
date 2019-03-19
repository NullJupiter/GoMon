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
	Log              bool
}

// Global variables.
var running = true
var config = getConfig()
var restartingCmd *exec.Cmd

func main() {
	restartingCmd = startCommand()

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
				if config.Log {
					log.Println("Modified file:", event.Name)
					log.Print("Restarting go program ...\n\n")
				}

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

	var workingDir string

	// Get flags
	command := flag.String("cmd", "", "Provide a command to execute and restart. If nothing is set, this defaults to \"go run $path/*.go\"")
	recursive := flag.Bool("r", true, "Search through the working directory recursively for file changes (set to true or false")
	quiet := flag.Bool("q", false, "Be quiet. Do not output anything to the standard output. (Errors are still displayed.)")
	flag.Parse()

	directories := flag.Args()

	if len(directories) > 0 {
		workingDir = directories[0]
	}

	// Output information if not set to quiet
	config.Log = !*quiet

	// Check if parameters are empty
	if workingDir == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of gomon:\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  gomon directory [directory...]\n\n")
		fmt.Fprint(flag.CommandLine.Output(), "  The given directories will be watched. In case no '-cmd' is given, the .go files in the first directory will be executed with 'go run'\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *command == "" {
		files, err := filepath.Glob(filepath.Join(workingDir, "*.go"))
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
		for _, dir := range directories {
			findDirectories(dir, &config.WatchDirectories)
		}
	} else {
		for _, dir := range directories {
			config.WatchDirectories = append(config.WatchDirectories, dir)
		}
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

	if config.Log {
		fmt.Println("Signal caught, Killing process", restartingCmd.Process.Pid)
	}

	err := syscall.Kill(-restartingCmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		killOnSignal()
	}

	restartingCmd.Wait()

	os.Exit(0)
}
