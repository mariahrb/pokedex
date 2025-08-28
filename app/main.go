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
	callback    func(*config) error
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

		if err := cmd.callback(cfg); err != nil {
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

func commandExit(cfg *config) error {
	fmt.Println("Closing Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
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

func commandMap(cfg *config) error {
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

func commandMapBack(cfg *config) error {
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
