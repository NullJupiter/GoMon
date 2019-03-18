# Gomon

## Not finished yet ! ! !

### Description
This is a program which allows you to push your productivity progress by monitoring your current working directory and restarting the main.go file on changes. It is something similar like nodemon for NodeJS.

### Usage
You need to provide the binary a "p" flag with the full path to the directory you want to monitor.
##### Example:
`./gomon -p /Users/mark/go/src/github.com/NullJupiter/GoTodoApp`

### What I learned
- how to use the fsnotify package
- how to program with flags and parse them
- how file monitoring works in general
