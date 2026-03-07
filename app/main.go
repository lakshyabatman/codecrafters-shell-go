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

var buildInCommands = []string{"echo", "exit", "type"}
var pathValue string = os.Getenv("PATH")

func main() {
	for {
		command := controlCommand()

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

func controlCommand() []string {
	fmt.Print("$ ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		panic("input failed")
	}
	return strings.Fields(line)
}
