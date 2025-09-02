package extentions_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mariahrb/pokedex/extentions"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{"https://example.com", []byte("testdata")},
		{"https://example.com/path", []byte("moretestdata")},
	}

	for i, cse := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := extentions.NewCache(interval)
			cache.Add(cse.key, cse.val)

			val, ok := cache.Get(cse.key)
			if !ok {
				t.Errorf("expected to find key %s", cse.key)
			}
			if string(val) != string(cse.val) {
				t.Errorf("expected value %s, got %s", cse.val, val)
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 20 * time.Millisecond
	const waitTime = baseTime + 30*time.Millisecond

	cache := extentions.NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")
	if !ok {
		t.Errorf("expected to find key initially")
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")
	if ok {
		t.Errorf("expected entry to be reaped")
	}
}
