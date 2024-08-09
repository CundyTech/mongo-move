package main

import (
	"fmt"
	"os"

	"github.com/square/exit"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	// Load and validate config
	config, err := load()
	if err != nil {
		fmt.Println(err)
		os.Exit(exit.FromError(err))
	}
	err = config.validate()
	if err != nil {
		fmt.Println(err)
		os.Exit(exit.FromError(err))
	}

	// Set up storage
	var s = NewStorage(config.Target, config.Source)

	// Load UI
	initialModel := model{
		Databases:         []string{},
		DatabaseChoice:    0,
		DatabaseChosen:    false,
		Collections:       []string{},
		CurrentCollection: 0,
		CollectionChoices: []collectionChoice{},
		CollectionsChosen: false,
		Quitting:          false,
		Storage:           s,
	}
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
