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
	Directories      []string
	Command          string
	CommandArguments []string
	Log              bool
	Recursive        bool
}

// Global variables.
var running = true
var config = getConfig()
var restartingCmd *exec.Cmd
var watchDirectories = make([]string, 0)
var watcher *fsnotify.Watcher

func main() {
	// Start the process for the first time
	restartingCmd = startCommand()

	// Listen to interrupt, terminate and kill signal to kill the process and exit
	go killOnSignal()

	refreshWatcher()

	// Start monitoring files
	go watchFiles()

	// Wait forever
	done := make(chan bool)
	<-done

	watcher.Close()
}

func refreshWatcher() {
	if watcher != nil {
		watcher.Close()
	}

	// Create file watcher
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if config.Recursive {
		// Recursively search for folders
		for _, dir := range config.Directories {
			findDirectories(dir, &watchDirectories)
		}
	} else {
		for _, dir := range config.Directories {
			watchDirectories = append(watchDirectories, dir)
		}
	}

	for _, dir := range watchDirectories {
		err := watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func watchFiles() {
	// Monitor files until running is set to false
	for running {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				file, err := os.Stat(event.Name)
				if err == nil && file.IsDir() {
					watcher.Add(event.Name) // TODO: Recursive?
				}
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				file, err := os.Stat(event.Name)
				if err == nil && file.IsDir() {
					refreshWatcher()
				}

			}

			// Catch write event (when file is saved/changed)
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
				if config.Log {
					log.Println("Modified file:", event.Name)
					log.Print("Restarting go program ...\n\n")
				}
				// Restart process
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
	// Search for directories in working directory an append them to the directories variable slice
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

	arguments := flag.Args()

	config.Recursive = *recursive

	config.Directories = make([]string, 0)
	cmdArgs := make([]string, 0)

	targetSlice := &config.Directories
	for _, arg := range arguments {
		if arg == "--" {
			targetSlice = &cmdArgs
			continue
		}
		*targetSlice = append(*targetSlice, arg)
	}

	// Add the first argument od flag.Args() corresponding to the working directory
	if len(config.Directories) > 0 {
		workingDir = config.Directories[0]
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

	// Get all go files in directory
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
	if len(cmdArgs) > 0 {
		config.CommandArguments = append(config.CommandArguments, cmdArgs...)
	}

	return config
}

func startCommand() *exec.Cmd {
	// Put the command together
	restartingCmd := exec.Command(config.Command, config.CommandArguments...)
	restartingCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	restartingCmd.Stdout = os.Stdout
	restartingCmd.Stderr = os.Stderr
	// Start the process
	err := restartingCmd.Start()
	if err != nil {
		log.Fatal("Something went wrong!")
	}

	return restartingCmd
}

func killOnSignal() {
	// Create channel to be notified when gomon is being interrupted, terminated or killed
	chanSigInt := make(chan os.Signal)
	signal.Notify(chanSigInt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// Block until channel is notified
	<-chanSigInt
	running = false

	if config.Log {
		fmt.Println("Signal caught, Killing process", restartingCmd.Process.Pid)
	}

	// Kill background proccess
	err := syscall.Kill(-restartingCmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		// If killing the proccess try it again and wait for next interrupt, termination or kill
		killOnSignal()
	}
	// Wait until process is actually killed
	restartingCmd.Wait()

	os.Exit(0)
}
