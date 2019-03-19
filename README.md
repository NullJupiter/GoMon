# GoMon - A file monitoring tool for Go/Golang

## Not finished yet ! ! !

### Description
This is a program which allows you to push your productivity progress by monitoring your working directory and restarting the go files on changes. It is something similar like nodemon for NodeJS.

### Usage


You need to provide the binary a `-p` flag with the full path to the directory you want to monitor.

##### Example:
`./gomon -p $(pwd)`



There are also two optional flags: `-cmd` and `-r`.



The `cmd` flag is used to run another program for example which executes some other stuff and then calls `go run *.go`. The `cmd` flag is set to `go run $path/*.go`.

##### run.sh:
`echo "Starting some stuff before entering the program...";`
`go run *.go;`
##### Example:
`./gomon -p $(pwd) -cmd "bash run.sh"`



The `r` flag is can be set to true or false. When set to `true` (by default) GoMon will recursively search through the working directory and its subdirectories for changes on files. Set to `false` GoMon will just detect changes in the working directory and will not monitor subdirectories.

##### Example:
`./gomon -p $(pwd) -r=false`


##### Command Line Agruments

 - `-p` - The path to the go files to be watched, run and restarted
 - `-r` - Watch the folder given with the `p`-flag recursively
 - `-cmd` - [Optional] The program to be run - in case it should different from the go files given with `-p`
 - `-q` - Do not output anything to the console

##### Test Programs

Test programs to verify how GoMon works can be found in the folders [/test/loop](/test/loop) and [/test/once](/test/once).

 - `/test/loop` Prints a "!" when started and then a "." every second until the process is stopped
 - `/test/once` Prints a message, then waits 5 seconds and exits

Run the test programs from the `src` folder like this:

 - `go run main.go -p ../test/loop/`
 - `go run main.go -p ../test/once/`


### ToDos and known bugs

 - The recursive folder watcher does not detect new folders.


### What I learned
- how to use the fsnotify package
- how to program with flags and parse them
- how file monitoring works in general
