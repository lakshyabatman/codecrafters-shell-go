package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var buildInCommands = []string{"echo", "exit", "type", "pwd"}
var pathValue string = os.Getenv("PATH")

func main() {
	for {
		command := parseCommand()

		if command[0] == "exit" {
			break
		} else {
			handleCommand(command)
		}
	}
}
func handleCommand(command []string) {
	if command[0] == "echo" {
		fmt.Println(strings.Join(command[1:], " "))
	} else if command[0] == "type" {
		handleTypeCommand(command)
	} else if command[0] == "pwd" {
		currentPath, _ := os.Getwd()
		fmt.Println(currentPath)
	} else if command[0] == "cd" {
		pathToGo := command[1]
		if command[1] == "~" {
			pathToGo = os.Getenv("HOME")
		}
		err := os.Chdir(pathToGo)
		if err != nil {
			fmt.Println(pathToGo + ": No such file or directory")
		}
	} else {
		path := checkAndGetInPaths(command[0], strings.Split(pathValue, ":"))
		if path == "" {
			fmt.Println(command[0] + ": not found")
		} else {
			var cmd *exec.Cmd
			if len(command) == 1 {
				cmd = exec.Command(command[0])
			} else {
				cmd = exec.Command(command[0], command[1:]...)
			}
			stdout, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Print(string(stdout))
		}
	}
}

func handleTypeCommand(command []string) {
	if slices.Contains(buildInCommands, command[1]) {
		fmt.Println(command[1] + " is a shell builtin")
	} else {
		path := checkAndGetInPaths(command[1], strings.Split(pathValue, ":"))
		if path == "" {
			fmt.Println(command[1] + ": not found")
		} else {
			fmt.Println(command[1] + " is " + path)
		}
	}
}

func checkAndGetInPaths(command string, paths []string) string {
	for _, path := range paths {
		foundPath, err := exec.LookPath(path + "/" + command)
		if err == nil {
			return foundPath
		}
	}
	return ""
}

func parseCommand() []string {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		panic("input failed")
	}
	var res []string
	var currentWord strings.Builder
	isSingleQuotes := false
	isDoubleQuotes := false
	isEscapeMode := false
	for _, ch := range strings.Trim(line, "\n") {
		// fmt.P/rintln(res)
		switch {
		case isEscapeMode:
			isEscapeMode = false
			currentWord.WriteRune(ch)
		case ch == '\\' && !isSingleQuotes:
			isEscapeMode = true
		case ch == '\'' && !isDoubleQuotes:
			isSingleQuotes = !isSingleQuotes

		case ch == '"' && !isSingleQuotes:
			isDoubleQuotes = !isDoubleQuotes
		case ch == ' ' && !isSingleQuotes && !isDoubleQuotes:
			if currentWord.Len() > 0 {
				res = append(res, currentWord.String())
				currentWord.Reset()
			}
		default:
			currentWord.WriteRune(ch)
		}
	}
	if currentWord.Len() > 0 {
		res = append(res, currentWord.String())
	}
	// fmt.Println(res)
	return res
}
