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

	var cctvm collectionChoicesViewModel
	cctvm.rowCount = 10
	cctvm.pageSize = 5
	cctvm.currentTableIndex = 0
	cctvm.sourceTable = buildTable(sourceCollectionsColumnName).WithPageSize(cctvm.pageSize).Focused(true).SortByAsc(sourceCollectionsColumnName)
	cctvm.targetTable = buildTable(targetCollectionsColumnName).WithPageSize(cctvm.pageSize).Focused(false).SortByAsc(targetCollectionsColumnName)
	cctvm.copyTaskTable = buildTable(sourceCollectionsColumnName, targetCollectionsColumnName).WithPageSize(5).Focused(false)
	cctvm.copyTasks = []collectionCopyTask{}
	cctvm.currentCopyTask = collectionCopyTask{}

	var dcvm databaseChoicesViewModel
	dcvm.currentTableIndex = 0
	dcvm.sourceDatabases = []string{}
	dcvm.sourceDatabaseChoice = ""
	dcvm.databasesChosen = false
	dcvm.sourceCollections = []string{}
	dcvm.sourceCurrentCollection = 0
	dcvm.sourcePageSize = 5
	dcvm.sourceTable = buildTable(sourceDatabasesColumnName).WithPageSize(dcvm.sourcePageSize).Focused(true).SortByAsc(sourceDatabasesColumnName)
	dcvm.targetDatabases = []string{}
	dcvm.targetDatabaseChoice = ""
	dcvm.targetCollections = []string{}
	dcvm.targetCurrentCollection = 0
	dcvm.targetPageSize = 5
	dcvm.targetTable = buildTable(targetDatabasesColumnName).WithPageSize(dcvm.targetPageSize).Focused(false).SortByAsc(targetDatabasesColumnName)
	dcvm.currentTableIndex = 0

	// Load terminal UI with intital model
	initialModel := model{
		databaseChoices:   dcvm,
		quitting:          false,
		storage:           s,
		collectionChoices: cctvm,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
