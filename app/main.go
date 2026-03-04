package main

import (
	"fmt"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	for {
		command := controlCommand()
	
       		if command == "exit" {
			break		
		}
		fmt.Println(command + ": command not found")
	
	}
}

func controlCommand() string {
	fmt.Print("$ ")
	var command string;
	fmt.Scanln(&command);
	return command
}

