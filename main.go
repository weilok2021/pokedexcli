package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

type cliCommand struct {
	name string
	description string
	callback func() error
}

var commandList map[string]cliCommand

func init() {
    commandList = map[string]cliCommand{
        "exit": {
            name:        "exit",
            description: "Exit the Pokedex",
            callback:    commandExit,
        },
        "help": {
            name:        "help",
            description: "Displays a help message",
            callback:    commandHelp,
        },
    }
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		// run exit callback function to terminate REPL
		if command, ok := commandList[input]; !ok {
			fmt.Println("Unknown Command")
		} else {
			err := command.callback()
			if err != nil {
				fmt.Errorf("%v", err)
			}
		}

	}
}


func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text) 
	words := strings.Fields(strings.ToLower(trimmed))
	// for _, word := range words {
	// 	fmt.Println(word)
	// }
	return words
} 


func commandExit() error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}


func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage: ")
	for command, _ := range commandList {
		fmt.Println("\t" + command)
	}
	return nil
}

