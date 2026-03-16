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
var dividerCommands = []string{">", "1>", "2>", ">>", "1>>", "2>>"}
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
		var commandError error
		if tokens[0] == "exit" {
			break
		} else {
			res, commandError = handleCommand(execTokens)
		}
		errorOutput := parseError(commandError)
		switch redirectionType {
		case "redirect":
			if err := os.WriteFile(outStd, []byte(res), 0644); err != nil {
				fmt.Errorf(err.Error())
			}
			if errorOutput != "" {
				fmt.Println(errorOutput)
			}
		case "redirectError":
			os.WriteFile(outStd, []byte(errorOutput), 0644)
			if res != "" {
				fmt.Println(res)
			}
		case "redirectAppend":
			f, err := os.OpenFile(outStd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			defer f.Close()
			if res != "" {
				if _, err = f.WriteString(fmt.Sprintf("\n%v", res)); err != nil {
					fmt.Printf("%v\n", err)
				}
			}

			if errorOutput != "" {
				fmt.Println(errorOutput)
			}
		case "redirectAppendError":
			f, err := os.OpenFile(outStd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			defer f.Close()

			if _, err = f.WriteString(fmt.Sprintf("%v", errorOutput)); err != nil {
				fmt.Printf("%v\n", err)
			}
			if res != "" {
				fmt.Println(res)
			}
		default:
			if res != "" {
				fmt.Println(res)
			}
		}

	}
}
func handleCommand(command []string) (string, error) {
	if command[0] == "echo" {
		return strings.Join(command[1:], " "), nil
	} else if command[0] == "type" {
		return handleTypeCommand(command), nil
	} else if command[0] == "pwd" {
		currentPath, _ := os.Getwd()
		return currentPath, nil
	} else if command[0] == "cd" {
		pathToGo := command[1]
		if command[1] == "~" {
			pathToGo = os.Getenv("HOME")
		}
		err := os.Chdir(pathToGo)
		if err != nil {
			return pathToGo + ": No such file or directory", err
		}
	} else {
		path := checkAndGetInPaths(command[0], strings.Split(pathValue, ":"))
		if path == "" {
			return command[0] + ": not found", nil
		} else {
			var cmd *exec.Cmd
			if len(command) == 1 {
				cmd = exec.Command(command[0])
			} else {
				cmd = exec.Command(command[0], command[1:]...)
			}
			stdout, err := cmd.Output()
			return string(strings.Trim(string(stdout), "\n")), err
		}
	}
	return "", nil
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
		var commandType string
		if t == ">" {
			commandType = "redirect"
		}
		switch {
		case t == ">" || t == "1>":
			commandType = "redirect"
		case t == "2>":
			commandType = "redirectError"

		case t == ">>" || t == "1>>":
			commandType = "redirectAppend"

		case t == "2>>":
			commandType = "redirectAppendError"
		}
		return tokens[:i], strings.Join(tokens[i+1:], " "), commandType, true

	}
	return tokens, "", "none", true
}

func parseError(err error) string {
	if err == nil {
		return ""
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return strings.Trim(string(exitErr.Stderr), "\n")
	}
	return fmt.Sprintf("%v", err)
}
