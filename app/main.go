package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mariahrb/pokedex/extentions"
)

// I want endpoint https://pokeapi.co/api/v2/location-area?limit=20, the ?limit is optional paramenter

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	next     *string
	previous *string
}

type locationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
}

type ExploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

var commands map[string]cliCommand
var pokeCache = extentions.NewCache(30 * time.Second)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	cfg := &config{}

	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Explore the Pokemon world forward",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Explore the Pokemon world backward",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explore a specific location area",
			callback:    commandExplore,
		},
	}

	for {
		fmt.Printf("Pokedex > ")

		if !scanner.Scan() {
			break
		}

		text := scanner.Text()

		words, err := cleanInput(text)
		if err != nil {
			fmt.Printf("Error cleaning input: %v\n", err)
			continue
		}

		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		cmd, ok := commands[commandName]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(cfg, words[1:]); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func cleanInput(text string) ([]string, error) {

	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", " ")
	parts := strings.Fields(text)

	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}

	return parts, nil
}

func commandExit(cfg *config, _ []string) error {
	fmt.Println("Closing Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage: \n\n")
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func fetchLocationAreas(url string) (*locationAreaResponse, error) {
	// Check cache	``
	if val, ok := pokeCache.Get(url); ok {
		fmt.Println("Cache hit:", url)
		var data locationAreaResponse
		if err := json.Unmarshal(val, &data); err != nil {
			return nil, fmt.Errorf("getting data %w", err)
		}
		return &data, nil
	}

	// Otherwise, fetch for API
	fmt.Println("Cache miss:", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting location response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body %w", err)
	}

	// Save response in cache
	pokeCache.Add(url, body)

	var data locationAreaResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("getting data %w", err)
	}
	return &data, nil
}

func commandMap(cfg *config, args []string) error {
	url := "https://pokeapi.co/api/v2/location-area?limit=20"
	if cfg.next != nil {
		url = *cfg.next
	}

	data, err := fetchLocationAreas(url)
	if err != nil {
		return fmt.Errorf("getting location data: %w", err)
	}

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	cfg.next = data.Next
	cfg.previous = data.Previous

	return nil
}

func commandMapBack(cfg *config, args []string) error {
	if cfg.previous == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	data, err := fetchLocationAreas(*cfg.previous)
	if err != nil {
		return fmt.Errorf("getting location data: %w", err)
	}

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	cfg.next = data.Next
	cfg.previous = data.Previous

	return nil
}

func commandExplore(cfg *config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: explore <location-area>")
	}
	areaName := args[0]

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)

	// First check cache
	if data, ok := pokeCache.Get(url); ok {
		return printExplore(areaName, data)
	}

	// Otherwise, fetch from API
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Save in cache
	pokeCache.Add(url, body)

	return printExplore(areaName, body)
}

func printExplore(areaName string, body []byte) error {
	var explore ExploreResponse
	if err := json.Unmarshal(body, &explore); err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", areaName)
	fmt.Println("Found Pokemon:")
	for _, encounter := range explore.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}
