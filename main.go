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
	callback func(pg *pagination) error
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
            description: "Displays next 20 location areas",
            callback:    commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "Display previous 20 location areas",
			callback: commandMapb,
		},
    }
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	pg := pagination{}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		// run exit callback function to terminate REPL
		if command, ok := commandList[input]; !ok {
			fmt.Println("Unknown Command")
		} else {
			err := command.callback(&pg)
			if err != nil {
				fmt.Errorf("%v", err)
			}
		}

	}
}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text) 
	words := strings.Fields(strings.ToLower(trimmed))
	return words
} 

func commandExit(pg *pagination) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp(pg *pagination) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage: ")
	for command, _ := range commandList {
		fmt.Println("\t" + command)
	}
	return nil
}

func commandMap(pg *pagination) error {
	var url string
	if pg.Next == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = pg.Next
	}
	res, err := http.Get(url)
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
	// Decode the json bytes into go struct
	var locations locationList
	if err := json.Unmarshal(body, &locations); err != nil {
		return err
	}

	// Display the 20 locations from map.results
	for _, locationArea := range locations.Results {
		fmt.Println(locationArea.Name)
	}

	// Update next field to return next page for next web response
	pg.Next = locations.Next
	pg.Previous = locations.Previous
	return nil
}

func commandMapb(pg* pagination) error {
	var url string
	if pg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	} else {
		url = pg.Previous
	}

	res, err := http.Get(url)
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
	// Decode the json bytes into go struct
	var locations locationList
	if err := json.Unmarshal(body, &locations); err != nil {
		return err
	}

	// Display the 20 locations from map.results
	for _, locationArea := range locations.Results {
		fmt.Println(locationArea.Name)
	}

	// update Previous field for return previous page for next web response
	pg.Previous = locations.Previous
	pg.Next = locations.Next
	return nil
}