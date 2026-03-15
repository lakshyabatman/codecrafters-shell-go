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
var dividerCommands = []string{">", "1>"}
var pathValue string = os.Getenv("PATH")

func main() {
	for {
		fmt.Print("$ ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			panic("input failed")
		}
		tokens := parseCommand(line)
		execTokens, outStd, redirectionType, _ := extractPipelineCommands(tokens)

		var res string
		if tokens[0] == "exit" {
			break
		} else {
			res = handleCommand(execTokens)
		}
		switch redirectionType {
		case "redirect":
			if err := os.WriteFile(outStd, []byte(res), 0644); err != nil {
				fmt.Errorf(err.Error())
			}
		default:
			fmt.Print(res)
		}
	}
}
func handleCommand(command []string) string {
	if command[0] == "echo" {
		return strings.Join(command[1:], " ")
	} else if command[0] == "type" {
		return handleTypeCommand(command)
	} else if command[0] == "pwd" {
		currentPath, _ := os.Getwd()
		return currentPath
	} else if command[0] == "cd" {
		pathToGo := command[1]
		if command[1] == "~" {
			pathToGo = os.Getenv("HOME")
		}
		err := os.Chdir(pathToGo)
		if err != nil {
			return pathToGo + ": No such file or directory"
		}
	} else {
		path := checkAndGetInPaths(command[0], strings.Split(pathValue, ":"))
		if path == "" {
			return command[0] + ": not found"
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
			return string(stdout)
		}
	}
	return ""
}

func handleTypeCommand(command []string) string {
	if slices.Contains(buildInCommands, command[1]) {
		return command[1] + " is a shell builtin"
	} else {
		path := checkAndGetInPaths(command[1], strings.Split(pathValue, ":"))
		if path == "" {
			return command[1] + ": not found"
		} else {
			return command[1] + " is " + path
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

func parseCommand(line string) []string {

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

func extractPipelineCommands(tokens []string) ([]string, string, string, bool) {
	for i, t := range tokens {
		if !slices.Contains(dividerCommands, t) {
			continue
		}
		if i+1 >= len(tokens) {
			return nil, "", "redirect", false
		}
		return tokens[:i], strings.Join(tokens[i+1:], " "), "redirect", true
	}
	return tokens, "", "none", true
}
