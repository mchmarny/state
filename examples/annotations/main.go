package main

import (
	"fmt"
	"log"

	"github.com/mchmarny/state/manager"
)

// Example is a simple struct with state annotations
type Example struct {
	Text   string `state:"text"`
	Number int    `state:"number"`
	Bool   bool   `state:"bool"`
	Other  string // this field will not be saved
}

func main() {
	// create a new state manager with file path and serialization type
	m, err := manager.NewStateManager(
		manager.WithFilePath("example.state"),
		manager.WithSerializationType(manager.STATE),
	)
	if err != nil {
		fmt.Printf("failed to create state manager: %v\n", err)
		return
	}

	// create a new instance of Example struct to hold state
	in := &Example{
		Text:   "Hello, World!",
		Number: 42,
		Bool:   true,
		Other:  "This will not be saved",
	}

	// save the state
	if err := m.Save(in); err != nil {
		log.Fatalf("failed to save state: %v", err)
	}

	// load the state
	var out Example
	if err := m.Load(&out); err != nil {
		log.Fatalf("failed to load state: %v", err)
	}

	// print the results
	fmt.Printf("Saved:  %+v\n", in)
	fmt.Printf("Loaded: %+v\n", &out)
}
