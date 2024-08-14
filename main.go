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
	var s = newStorage(config.Target, config.Source)

	var cctvm collectionChoiceTableViewModel
	cctvm.rowCount = 10
	cctvm.pageSize = 5
	cctvm.currentTableIndex = 0
	cctvm.sourceTable = buildTable(sourceColumnName).WithPageSize(cctvm.pageSize).Focused(true)
	cctvm.targetTable = buildTable(targetColumnName).WithPageSize(cctvm.pageSize).Focused(false)
	cctvm.copyTaskTable = buildTable(sourceColumnName, targetColumnName).WithPageSize(5).Focused(false)
	cctvm.copyTasks = []collectionCopyTask{}
	cctvm.currentCopyTask = collectionCopyTask{}

	var dscvm databaseSourceChoicesViewModel
	dscvm.sourceDatabases = []string{}
	dscvm.sourceDatabaseChoice = 0
	dscvm.sourceDatabaseChosen = false
	dscvm.sourceCollections = []string{}
	dscvm.sourceCurrentCollection = 0

	var dtcvm databaseTargetChoicesViewModel
	dtcvm.targetDatabases = []string{}
	dtcvm.targetDatabaseChoice = 0
	dtcvm.targetDatabaseChosen = false
	dtcvm.targetCollections = []string{}
	dtcvm.targetCurrentCollection = 0

	// Load terminal UI with intital model
	initialModel := model{
		dscvm:    dscvm,
		dtcvm:    dtcvm,
		quitting: false,
		storage:  s,
		cctvm:    cctvm,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
