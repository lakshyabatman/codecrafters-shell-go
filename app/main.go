package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	"github.com/codecrafters-io/shell-starter-go/app/bellcompleter"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var buildInCommands = []string{"echo", "exit", "type", "pwd", "history"}
var dividerCommands = []string{">", "1>", "2>", ">>", "1>>", "2>>"}
var pathValue string = os.Getenv("HOME") + "/.shell_history/" + strconv.Itoa(os.Getpid())

func main() {
	// fmt.Println(pathValue)
	f, _ := os.OpenFile(pathValue, os.O_CREATE, 0600)
	f.Close()
	prefixCompleter := readline.NewPrefixCompleter(
		readline.PcItem("echo"),
		readline.PcItem("exit"),
	)
	completer := &bellcompleter.BellCompleter{Completer: prefixCompleter, TabCount: 0}
	cfg := &readline.Config{
		Prompt:                 "$ ",
		AutoComplete:           completer,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		HistoryFile:            pathValue,
		DisableAutoSaveHistory: false,
	}

	// Initialize the readline instance.
	rl, err := readline.NewEx(cfg)
	if err != nil {
		panic(err)
	}
	defer rl.Close() // Ensure the terminal is restored to its original state on exit

	for {

		if err != nil {
			panic("input failed")
		}
		line, err := rl.Readline()
		tokens := parseCommand(line)
		commands := make([][]string, 1)
		i := 0
		for _, token := range tokens {
			if token == "|" {
				// fmt.Println("Ci")
				commands = append(commands, []string{})
				i++
			} else {
				commands[i] = append(commands[i], token)
			}

		}
		if err != nil {
			panic("Failed!")
		}
		var wg sync.WaitGroup
		pipeReaders := make([]*io.PipeReader, len(commands))
		pipeWriters := make([]*io.PipeWriter, len(commands))
		for i := 0; i < len(commands)-1; i++ {
			pipeReaders[i+1], pipeWriters[i] = io.Pipe()
		}
		for i, command := range commands {
			if command[0] == "exit" {
				os.Exit(0)
			}
			wg.Add(1)
			go func(cmd []string, idx int) {
				if pipeReaders[idx] != nil {
					defer pipeReaders[idx].Close()
				}
				if pipeWriters[idx] != nil {
					defer pipeWriters[idx].Close()
				}
				executeCommand(cmd, pipeWriters[idx], pipeReaders[idx])
				defer wg.Done()

			}(command, i)
		}
		wg.Wait()

	}
}

func executeCommand(command []string, pipeWriter *io.PipeWriter, pipeReader *io.PipeReader) {
	execTokens, outStd, redirectionType, _ := extractPipelineCommands(command)

	stdin, stdout, stderr := getIOs(pipeReader, pipeWriter, redirectionType, outStd)

	var commandError error
	if command[0] == "echo" {
		fmt.Fprintln(stdout, strings.Join(execTokens[1:], " "))
	} else if command[0] == "type" {
		fmt.Fprintln(stdout, handleTypeCommand(command))
	} else if command[0] == "pwd" {
		currentPath, _ := os.Getwd()
		fmt.Fprintln(stdout, currentPath)
	} else if command[0] == "cd" {
		pathToGo := command[1]
		if command[1] == "~" {
			pathToGo = os.Getenv("HOME")
		}
		if err := os.Chdir(pathToGo); err != nil {
			fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", pathToGo)
		}
	} else if command[0] == "history" {
		f, _ := os.ReadFile(pathValue)
		lines := strings.Split(strings.TrimRight(string(f), "\n"), "\n")
		linesToPrint := len(lines)
		if len(execTokens) > 1 {

			i, _ := strconv.Atoi(execTokens[1])
			linesToPrint = min(linesToPrint, i)
			// fmt.Println(linesToPrint)
		}
		for i, l := range lines[(len(lines) - linesToPrint):] {
			if l != "" {
				fmt.Fprintf(stdout, "  %d  %s\n", (i + (len(lines) - linesToPrint) + 1), l)
			}
		}
	} else {
		commandError = executeSingleCommand(execTokens, stdin, stdout, stderr)

	}
	errorOutput := parseError(commandError)
	switch redirectionType {
	case "redirect", "redirectAppend":
		if errorOutput != "" {
			fmt.Println(errorOutput)
		}
	case "redirectError", "redirectAppendError":
		if stdout != nil {

		}
	}
}

func getIOs(pipeReader *io.PipeReader, pipeWriter *io.PipeWriter, redirectionType string, outStd string) (io.Reader, io.Writer, io.Writer) {
	var stdin io.Reader
	if pipeReader != nil {
		stdin = pipeReader
	}
	var stdout io.Writer
	var stderr io.Writer
	stdout = os.Stdout
	stderr = os.Stderr
	if pipeWriter != nil {
		stdout = pipeWriter
	} else {
		switch redirectionType {
		case "redirect":
			f, _ := os.OpenFile(outStd, os.O_CREATE|os.O_WRONLY, 0644)
			stdout = f
		case "redirectAppend":
			f, _ := os.OpenFile(outStd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			stdout = f
		case "redirectError":
			f, _ := os.OpenFile(outStd, os.O_CREATE|os.O_WRONLY, 0644)
			stderr = f
		case "redirectAppendError":
			f, _ := os.OpenFile(outStd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			stderr = f
		}
	}
	return stdin, stdout, stderr

}

func executeSingleCommand(command []string, pipeReader io.Reader, pipeWriter io.Writer, pipeErr io.Writer) error {

	path := checkAndGetInPaths(command[0], strings.Split(pathValue, ":"))
	if path == "" {
		fmt.Fprintln(os.Stderr, command[0]+": not found")
		return nil
	} else {
		var cmd *exec.Cmd
		if len(command) == 1 {
			cmd = exec.Command(command[0])
		} else {
			cmd = exec.Command(command[0], command[1:]...)
		}
		cmd.Stdin = pipeReader
		cmd.Stdout = pipeWriter
		cmd.Stderr = pipeErr
		return cmd.Run()
	}

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
