package main

import (
	"fmt"
	"log"

	"github.com/mchmarny/state/manager"
)

type Example struct {
	Text   string
	Number int
	Bool   bool
}

func main() {
	// Create a new state manager using the default file path (~/.state)
	m, err := manager.NewStateManager()
	if err != nil {
		fmt.Printf("failed to create state manager: %v\n", err)
		return
	}

	// Create a struct that holds the state
	in := &Example{
		Text:   "Hello, World!",
		Number: 42,
		Bool:   true,
	}

	// Save the state
	if err := m.Save(in); err != nil {
		log.Fatalf("failed to save state: %v", err)
	}

	// Load the state
	var out Example
	if err := m.Load(&out); err != nil {
		log.Fatalf("failed to load state: %v", err)
	}

	// Print the saved and loaded state
	fmt.Printf("Saved:  %+v\n", in)
	fmt.Printf("Loaded: %+v\n", &out)
}
