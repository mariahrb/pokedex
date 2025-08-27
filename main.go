package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands map[string]cliCommand

func main() {
	scanner := bufio.NewScanner(os.Stdin)

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

		if err := cmd.callback(); err != nil {
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

func commandExit() error {
	fmt.Println("Closing Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage; \n")
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}
