package main

import (
	"fmt"
	"os"
	"time"

	math "math/rand/v2"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	cctvm.sourceTable = buildTable(sourceCollectionsColumnName, recordsCountColumnName).
		WithPageSize(cctvm.pageSize).
		Focused(true).
		SortByAsc(sourceCollectionsColumnName)
	cctvm.targetTable = buildTable(targetCollectionsColumnName, recordsCountColumnName).
		WithPageSize(cctvm.pageSize).
		Focused(false).
		SortByAsc(targetCollectionsColumnName)
	cctvm.copyTaskTable = buildTable(sourceCollectionsColumnName, targetCollectionsColumnName, CopyStatusColumnName).
		WithPageSize(5).
		Focused(false)
	cctvm.copyTasks = []collectionCopyTask{}
	cctvm.currentCopyTask = collectionCopyTask{}
	cctvm.debounce = 2 * time.Second
	cctvm.altscreen = false

	var dcvm databaseChoicesViewModel
	dcvm.currentTableIndex = 0
	dcvm.sourceDatabases = []string{}
	dcvm.sourceDatabaseChoice = ""
	dcvm.databasesChosen = false
	dcvm.sourceCollections = []collection{}
	dcvm.sourceCurrentCollection = 0
	dcvm.sourcePageSize = 5
	dcvm.sourceTable = buildTable(sourceDatabasesColumnName).
		WithPageSize(dcvm.sourcePageSize).
		Focused(true).
		SortByAsc(sourceDatabasesColumnName)
	dcvm.targetDatabases = []string{}
	dcvm.targetDatabaseChoice = ""
	dcvm.targetCollections = []collection{}
	dcvm.targetCurrentCollection = 0
	dcvm.targetPageSize = 5
	dcvm.targetTable = buildTable(targetDatabasesColumnName).
		WithPageSize(dcvm.targetPageSize).
		Focused(false).
		SortByAsc(targetDatabasesColumnName)
	dcvm.currentTableIndex = 0
	dcvm.databasesLoaded = false
	dcvm.debounce = 2 * time.Second

	var keyModel keyModel
	keyModel.quitting = false
	keyModel.keys = keys
	keyModel.inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7"))

	// Available spinners
	spinners := []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	var sp = spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	sp.Spinner = spinners[math.IntN(8)]

	// Load terminal UI with intital model
	initialModel := model{
		databaseChoices:   dcvm,
		keyBindings:       keyModel,
		storage:           s,
		collectionChoices: cctvm,
		spinner:           sp,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
