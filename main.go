package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/square/exit"
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
		TargetDatabases:          []string{},
		SourceDatabases:          []string{},
		SourceDatabaseChoice:     0,
		TargetDatabaseChoice:     0,
		SourceDatabaseChosen:     false,
		TargetDatabaseChosen:     false,
		TargetCollections:        []string{},
		SourceCollections:        []string{},
		TargetCurrentCollection:  0,
		SourceCurrentCollection:  0,
		CollectionsChosen:        false,
		Quitting:                 false,
		Storage:                  s,
		SourceTable:              genTable(sourceColumnName).WithPageSize(5).Focused(true),
		TargetTable:              genTable(targetColumnName).WithPageSize(5).Focused(false),
		CopyTaskTable:            genTable(sourceColumnName, targetColumnName).WithPageSize(5).Focused(false),
		CopyTasks:                []collectionCopyTask{},
		CurrentCopyTask:          collectionCopyTask{},
		RowCount:                 10,
		CollectionViewTableIndex: 0,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
