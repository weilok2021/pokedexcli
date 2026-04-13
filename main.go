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
	"math/rand"
)

type cliCommand struct {
	name string
	description string
	callback func(pg *pagination, argument string) error
}

var commandList map[string]cliCommand
var pokemonStorage map[string]Pokemon
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
		"catch": {
			name: "catch",
			description: "Attempt to catch the pokemon based on name",
			callback: commandCatch,
		},
		"inspect": {
			name: "inspect",
			description: "Inspect the catched pokemon info in stats (if existed)",
			callback: commandInspect,
		},
    }

	cache = pokecache.NewCache(5 * time.Minute)
	pokemonStorage = make(map[string]Pokemon)
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
			var argv string // could be locationArea or pokemon name
			if len(input) == 2 {
				argv = input[1]
			}
			err := command.callback(&pg, argv)
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

func commandCatch(pg* pagination, pokemonName string) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", pokemonName)
	// Check to see if the url already stored in cache, 
	// if it is, get the pokemon and attempt to catch it.
	if bytes, ok := cache.Get(url); ok {
		// Decode the json bytes into go struct
		var pokemon Pokemon
		if err := json.Unmarshal(bytes, &pokemon); err != nil {
			return err
		}

		fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
		fmt.Printf("Reattempt to catch %s!!\n", pokemon.Name)

		if rand.Intn(pokemon.BaseExperience) <= 50 {
			pokemonStorage[pokemon.Name] = pokemon
			fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)
			fmt.Printf("%s was caught!\n", pokemon.Name)
		} else {
			fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)
			fmt.Printf("%s escaped!\n", pokemon.Name)
		}
		return nil
	}


	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// Convert *Response into slices of Byte
	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	// cache response bytes
	cache.Add(url, body)

	var pokemon Pokemon
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)

	if rand.Intn(pokemon.BaseExperience) <= 50 {
		fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)
		// keep track of catched pokemon
		pokemonStorage[pokemon.Name] = pokemon
		fmt.Printf("%s was caught!\n", pokemon.Name)
	} else {
		fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}
	return nil
}

func commandInspect(pg* pagination, pokemonName string) error {
	if pokemon, ok := pokemonStorage[pokemonName]; ok {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)

	} else {
		fmt.Println("You haven't catch this pokemon.")
	}
	return nil
}
