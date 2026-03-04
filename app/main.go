package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	for {
		command := controlCommand()
       		if command[0] == "exit" {
			break		
		} else if command[0] == "echo" {
			fmt.Println(strings.Join(command[1:], " "))
		}else {	
			fmt.Println(command[0] + ": command not found")
		}	
	}
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

