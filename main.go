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
	"time"
	"github.com/weilok2021/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name string
	description string
	callback func(pg *pagination, areaName string) error
}

var commandList map[string]cliCommand
var cache *pokecache.Cache

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
		"explore": {
			name: "explore",
			description: "Display all pokemons in this selected locationArea",
			callback: commandExplore,
		},
    }

	cache = pokecache.NewCache(5 * time.Minute)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	pg := pagination{}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := cleanInput(scanner.Text()) // Split command args into slices of strings
		// run exit callback function to terminate REPL
		if command, ok := commandList[input[0]]; !ok {
			fmt.Println("Unknown Command")
		} else {
			var areaExplore string 
			if input[0] == "explore" && len(input) == 2{
				areaExplore = input[1]
			}
			err := command.callback(&pg, areaExplore)
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

func commandExit(pg *pagination, areaName string) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandHelp(pg *pagination, areaName string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage: ")
	for command, _ := range commandList {
		fmt.Println("\t" + command)
	}
	return nil
}

func commandMap(pg *pagination, areaName string) error {
	var url string
	if pg.Next == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	} else {
		url = pg.Next
	}

	// check to see if the url already stored in cache, 
	// if it is, Print locationarea here and return.
	if bytes, ok := cache.Get(url); ok {
		// Decode the json bytes into go struct
		var locations locationList
		if err := json.Unmarshal(bytes, &locations); err != nil {
			return err
		}

		// Display the 20 locations from map.results
		for _, locationArea := range locations.Results {
			fmt.Println(locationArea.Name)
		}
		pg.Next = locations.Next
		pg.Previous = locations.Previous
		return nil
	}

	// returns a response pointer to Response struct
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

	// Cache this resposne bytes, so next read will be much faster
	cache.Add(url, body)

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

func commandMapb(pg* pagination, areaName string) error {
	var url string
	if pg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	} else {
		url = pg.Previous
	}

	// Check to see if the url already stored in cache, 
	// if it is, Print locationarea here and return.
	if bytes, ok := cache.Get(url); ok {
		// Decode the json bytes into go struct
		var locations locationList
		if err := json.Unmarshal(bytes, &locations); err != nil {
			return err
		}

		// Display the 20 locations from map.results
		for _, locationArea := range locations.Results {
			fmt.Println(locationArea.Name)
		}		

		pg.Next = locations.Next
		pg.Previous = locations.Previous
		return nil
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

	// Cache this response bytes, so next read will be much faster
	cache.Add(url, body)

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

func commandExplore(pg* pagination, areaName string) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)

	// Check to see if the url already stored in cache, 
	// if it is, Print all pokemons in this area then return.
	if bytes, ok := cache.Get(url); ok {
		// Decode the json bytes into go struct
		var areaExplored locationAreaExplore
		if err := json.Unmarshal(bytes, &areaExplored); err != nil {
			return err
		}

		// "Display all pokemons in this locationArea",
		for _, pokemonStruct := range areaExplored.PokemonEncounters {
			fmt.Println(pokemonStruct.Pokemon.Name)
		}		
		return nil
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

	// Cache this response bytes, so next read will be much faster
	cache.Add(url, body)

	var areaExplored locationAreaExplore
	if err := json.Unmarshal(body, &areaExplored); err != nil {
		return err
	}

	// "Display all pokemons in this locationArea"
	for _, pokemonStruct := range areaExplored.PokemonEncounters {
		fmt.Println(pokemonStruct.Pokemon.Name)
	}		
	return nil
}