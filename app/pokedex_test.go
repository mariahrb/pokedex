package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"  pikachu  ", []string{"pikachu"}},
		{" charmander ", []string{"charmander"}},
		{"Bulbasaur, Squirtle", []string{"bulbasaur", "squirtle"}},
	}

	for _, c := range cases {
		actual, err := cleanInput(c.input)
		if err != nil {
			t.Errorf("unexpected error for input %q: %v", c.input, err)
			continue
		}

		if len(actual) != len(c.expected) {
			t.Errorf("For input %q, expected length %d, got %d. Expected: %v, Got: %v", c.input, len(c.expected), len(actual), c.expected, actual)
			continue
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("For input %q, expected %q at index %d, got %q", c.input, expectedWord, i, word)
			}
		}
	}
}
