# GoMon - A file monitoring tool for Go/Golang

## Not finished yet ! ! !
##### There is a bug that you can only run programs that never end until there is an interrupt like programs with a endless loop or HTTP Servers. Programs that give an output and then close themselves cannot be run by this program yet.

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



### What I learned
- how to use the fsnotify package
- how to program with flags and parse them
- how file monitoring works in general
