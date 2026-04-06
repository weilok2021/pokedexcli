package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"net/http"
	"encoding/json"
	"log"
	"io"
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
		"map": {
			name:        "map",
            description: "Displays next 20 locations",
            callback:    commandMap,
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

func commandMap() error {
	res, err := http.Get("https://pokeapi.co/api/v2/location")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	// Parse json into a slice of bytes
	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Debug: %v\n", res)
	// Decode the json bytes into go struct
	var locations locationList
	if err := json.Unmarshal(body, &locations); err != nil {
		return err
	}
	// Display the 20 locations from map.results
	fmt.Printf("Debug: %v", locations)
	fmt.Println("The first 20 locations are: ")
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	return nil
}