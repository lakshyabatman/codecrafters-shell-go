package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"slices"
	)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var supportedCommands = []string{"echo", "exit", "type"}


func main() {
	for {
		command := controlCommand()
		
       		if command[0] == "exit" {
			break		
		}else {
			handleCommand(command)
		}	
	}
}
func handleCommand(command []string) {
	if command[0] == "echo" {
		fmt.Println(strings.Join(command[1:], " "))
	} else if command[0] == "type" {
		if slices.Contains(supportedCommands, command[1]) {
			fmt.Println(command[1] + " is a shell builtin")			
		}else {
			fmt.Println(command[1] + ": command not found")
		}
	}else  {    
			fmt.Println(command[0] + ": command not found")
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

