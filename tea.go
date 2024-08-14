package main

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	targetColumnName   = "Target Collections"
	sourceColumnName   = "Source Collections"
	taskMapColumnsName = "Collections Map"
	progressBarWidth   = 71
	dotChar            = " • "
	banner             = `
  __  __                           __  __                
 |  \/  |                         |  \/  |               
 | \  / | ___  _ __   __ _  ___   | \  / | _____   _____ 
 | |\/| |/ _ \| '_ \ / _' |/ _ \  | |\/| |/ _ \ \ / / _ \
 | |  | | (_) | | | | (_| | (_) | | |  | | (_) \ V /  __/
 |_|  |_|\___/|_| |_|\__, |\___/  |_|  |_|\___/ \_/ \___|
                      __/ |                              
                     |___/                               
`
	errorImage = `                                             
                                             
      ::.                         .::.       
     *%%%:      :-=**#**=-:      .#%%*       
   .=%%%%*.  .=%%%%%%%%%%%%%=.  .*%%%%=.     
   *%%%%%%%:=%%%%%%%%%%%%%%%%%=:#%%%%%%*     
    :-.  =-=%%%%%%%%%%%%%%%%%%%=:=  .-:      
           #%%%%%%%%%%%%%%%%%%%%             
           %%%%%%%%%%%%%%%%%%%%%.            
           *%%%%%%%%%%%%%%%%%%%*             
           .%%+     #%%.    =%%:             
          ==:%*    +%%%+.   +%-==            
   =%%%##%%=-%%%%%%%*.+%%%%%%%==%%##%%%+     
   :#%%%%%-. -=+*%%%*++%%%*+==..-#%%%%%-     
     *%%%-       *%%%%%%%#       :%%%#       
     :==:        =*%%%%%*=        :==:       
                    ...                      
                                             
`
)

// General stuff for styling the view
var (
	bannerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#54ad48"))
	keywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
)

type copyMsg struct {
	collectionId int
}

type collections struct {
	target []string
	source []string
}

type TableData struct {
	Columns []string
	Rows    []map[string]interface{}
}

type collectionCopyTask struct {
	id       int
	target   string
	source   string
	spinner  spinner.Model
	complete bool
}

type (
	getSourceDatabasesMsg []string
	getTargetDatabasesMsg []string
	getCollectionsMsg     collections
)

type errMsg struct {
	err     error
	context string
}

type fatalError struct {
	text    string
	context string
}

// Model for view where user chooses the source and target collections
type collectionChoiceTableViewModel struct {
	sourceTable       table.Model          // Table that displays collections in the source database
	targetTable       table.Model          // Table that displays collections in the target database
	copyTaskTable     table.Model          // Table that displays chosen source and target collection maps
	copyTasks         []collectionCopyTask // Vollection of source and targets where data will be moved from and to respectively
	currentCopyTask   collectionCopyTask   // Vurrent user selection of source and target collections
	currentTableIndex int                  // Index of the table is currently in use by user. 0 = sourceTable, 1 = targetTable and 2 = copyTaskTable
	pageSize          int                  // Default size of a page of all tables
	rowCount          int                  // The amount of rows in a table
	collectionsChosen bool                 // Has user made collection choices
}

type databaseSourceChoicesViewModel struct {
	sourceDatabases         []string // Databases on server
	sourceDatabaseChoice    int      // Database chosen by user
	sourceDatabaseChosen    bool     // Has user made database selection
	sourceCollections       []string // Collections in database
	sourceCurrentCollection int      // Collection cursor is current on
}

type databaseTargetChoicesViewModel struct {
	targetDatabases         []string // Databases on server
	targetDatabaseChoice    int      // Database chosen by user
	targetDatabaseChosen    bool     // Has user made database selection
	targetCollections       []string // Collections in database
	targetCurrentCollection int      // Collection cursor is current on
}

// Main model
type model struct {
	quitting   bool                           // Has user quit application
	storage    storage                        // Storage
	fatalError *fatalError                    // Fatal Error details
	dscvm      databaseSourceChoicesViewModel // Model for databaseSourceChoicesView view
	dtcvm      databaseTargetChoicesViewModel // Model for databaseTargetChoicesView view
	cctvm      collectionChoiceTableViewModel // Model for collectionChoiceTable view
}

// Init function that returns an initial command for the application to run
func (m model) Init() tea.Cmd {
	return m.getSourceDatabases
}

// Commands -  Functions that perform some I/O and then return a Msg.
// https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands/

func (m model) getSourceDatabases() tea.Msg {
	databases, err := m.storage.getSourceDatabases()
	if err != nil {
		return errMsg{err, "getting source databases"}
	}

	return getSourceDatabasesMsg(databases)
}

func (m model) getTargetDatabases() tea.Msg {
	databases, err := m.storage.getTargetDatabases()
	if err != nil {
		return errMsg{err, "getting target databases"}
	}

	return getTargetDatabasesMsg(databases)
}

func (m model) getCollections() tea.Msg {
	var collections collections
	var err error

	collections.target, err = m.storage.getTargetCollections(m.dtcvm.targetDatabases[m.dtcvm.targetDatabaseChoice])
	if err != nil {
		return errMsg{err, "getting target collections"}
	}

	collections.source, err = m.storage.getSourceCollections(m.dscvm.sourceDatabases[m.dscvm.sourceDatabaseChoice])
	if err != nil {
		return errMsg{err, "getting source collections"}
	}

	return getCollectionsMsg(collections)
}

func (m model) copyData() []tea.Cmd {
	var err error
	var cmds []tea.Cmd

	for _, c := range m.cctvm.copyTasks {
		cmd := func() tea.Msg {
			err = m.storage.copy(c.source, c.target, m.dscvm.sourceDatabases[m.dscvm.sourceDatabaseChoice], m.dtcvm.targetDatabases[m.dtcvm.targetDatabaseChoice])
			if err != nil {
				return errMsg{err, "copying records"}
			}

			return copyMsg{collectionId: c.id}
		}

		cmds = append(cmds, cmd)
	}

	return cmds
}

// Updates - Functions that handle incoming events and updates the model accordingly
// https://github.com/charmbracelet/bubbletea#the-update-method

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Make sure these keys always quit
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case getSourceDatabasesMsg:
		m.dscvm.sourceDatabases = msg
		return m, tea.ClearScreen
	case getTargetDatabasesMsg:
		m.dtcvm.targetDatabases = msg
		return m, tea.ClearScreen
	case getCollectionsMsg:
		m.dscvm.sourceCollections = msg.source
		m.dtcvm.targetCollections = msg.target
		m.buildCollectionTableRows()
		return m, tea.ClearScreen
	case copyMsg:
		for i := 0; i < len(m.cctvm.copyTasks); i++ {
			if msg.collectionId == m.cctvm.copyTasks[i].id {
				m.cctvm.copyTasks[i].complete = true
			}
		}
		return m, nil
	case spinner.TickMsg:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)
		for i := 0; i < len(m.cctvm.copyTasks); i++ {
			m.cctvm.copyTasks[i].spinner, cmd = m.cctvm.copyTasks[i].spinner.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		return m, tea.Batch(cmds...)
	case errMsg:
		m.fatalError = &fatalError{text: msg.err.Error(), context: msg.context}

		return m, tea.ClearScreen
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.dscvm.sourceDatabaseChosen {
		return updateSourceDatabaseChoices(msg, m)
	} else if !m.dtcvm.targetDatabaseChosen {
		return updateTargetDatabaseChoices(msg, m)
	} else if !m.cctvm.collectionsChosen {
		return updateCollectionChoiceTable(msg, m)
	}

	return m, nil
}

// Update loop for the first view where you're choosing a database.
func updateSourceDatabaseChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.dscvm.sourceDatabaseChoice++
			if m.dscvm.sourceDatabaseChoice >= len(m.dscvm.sourceDatabases) {
				m.dscvm.sourceDatabaseChoice = len(m.dscvm.sourceDatabases) - 1
			}
		case "up":
			m.dscvm.sourceDatabaseChoice--
			if m.dscvm.sourceDatabaseChoice < 0 {
				m.dscvm.sourceDatabaseChoice = 0
			}
		case "enter":
			m.dscvm.sourceDatabaseChosen = true
			return m, m.getTargetDatabases
		}
	}

	return m, nil
}

// Update loop for the first view where you're choosing a database.
func updateTargetDatabaseChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.dtcvm.targetDatabaseChoice++
			if m.dtcvm.targetDatabaseChoice >= len(m.dtcvm.targetDatabases) {
				m.dtcvm.targetDatabaseChoice = len(m.dtcvm.targetDatabases) - 1
			}
		case "up":
			m.dtcvm.targetDatabaseChoice--
			if m.dtcvm.targetDatabaseChoice < 0 {
				m.dtcvm.targetDatabaseChoice = 0
			}
		case "enter":
			m.dtcvm.targetDatabaseChosen = true
			return m, m.getCollections
		}
	}

	return m, nil
}

func updateCollectionChoiceTable(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)

		case "a":
			m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(true)
			m.cctvm.targetTable = m.cctvm.targetTable.Focused(false)
			m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.Focused(false)

		case "b":
			m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(false)
			m.cctvm.targetTable = m.cctvm.targetTable.Focused(true)
			m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.Focused(false)

		case "enter":
			if len(m.cctvm.copyTasks) != 0 {
				m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(false)
				m.cctvm.targetTable = m.cctvm.targetTable.Focused(false)
				m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.Focused(true)
			}

		case "s":
			if len(m.cctvm.copyTasks) == 0 {
				return m, nil
			}

			m.cctvm.collectionsChosen = true

			var cmds []tea.Cmd
			var c = m.copyData()
			cmds = append(cmds, c...)

			for i := 0; i < len(m.cctvm.copyTasks); i++ {
				cmd := func() tea.Msg {
					return m.cctvm.copyTasks[i].spinner.Tick()
				}

				cmds = append(cmds, cmd)
			}

			return m, tea.Batch(cmds...)

		case "u":
			m.cctvm.sourceTable = m.cctvm.sourceTable.WithPageSize(m.cctvm.sourceTable.PageSize() - 1)
			m.cctvm.targetTable = m.cctvm.targetTable.WithPageSize(m.cctvm.targetTable.PageSize() - 1)

		case "i":
			m.cctvm.sourceTable = m.cctvm.sourceTable.WithPageSize(m.cctvm.sourceTable.PageSize() + 1)
			m.cctvm.targetTable = m.cctvm.targetTable.WithPageSize(m.cctvm.targetTable.PageSize() + 1)

		case "r":
			m.cctvm.sourceTable = m.cctvm.sourceTable.WithCurrentPage(rand.Intn(m.cctvm.sourceTable.MaxPages()) + 1)
			m.cctvm.targetTable = m.cctvm.targetTable.WithCurrentPage(rand.Intn(m.cctvm.targetTable.MaxPages()) + 1)

		case "z":
			if m.cctvm.rowCount < 10 {
				break
			}

			m.cctvm.rowCount -= 10
			m.buildCollectionTableRows()

		case "x":
			m.cctvm.rowCount += 10
			m.buildCollectionTableRows()

		case "delete":
			// Only delete form copy map table
			if m.cctvm.currentTableIndex == 2 {
				//Todo Delete mappings
			}

		case " ":
			if m.cctvm.sourceTable.GetFocused() {
				m.cctvm.currentTableIndex++
				row := m.cctvm.sourceTable.HighlightedRow()

				var value = row.Data[sourceColumnName].(string)
				m.cctvm.currentCopyTask.source = value

				// Delete collection so it can't be selected again
				var i = m.cctvm.sourceTable.GetHighlightedRowIndex()
				m.dscvm.sourceCollections = removeItem(m.dscvm.sourceCollections, i)
				m.buildCollectionTableRows()
				m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(false)
				m.cctvm.targetTable = m.cctvm.targetTable.Focused(true)

			} else if m.cctvm.targetTable.GetFocused() {
				m.cctvm.currentTableIndex++
				row := m.cctvm.targetTable.HighlightedRow()
				var value = row.Data[targetColumnName].(string)
				m.cctvm.currentCopyTask.target = value

				// Delete collection so it can't be selected again
				var i = m.cctvm.targetTable.GetHighlightedRowIndex()
				m.dtcvm.targetCollections = removeItem(m.dtcvm.targetCollections, i)
				m.buildCollectionTableRows()
				m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(true)
				m.cctvm.targetTable = m.cctvm.targetTable.Focused(false)
			}

			// User has chose target and source, so add to copy map
			if m.cctvm.currentTableIndex == 2 {
				m.cctvm.copyTasks = append(m.cctvm.copyTasks, m.cctvm.currentCopyTask)
				m.buildCollectionMapRows()
				m.cctvm.currentTableIndex = 0
			}

			// No more viable copy maps to be selected
			if len(m.dscvm.sourceCollections) == 0 && len(m.dtcvm.targetCollections) == 0 {
				m.cctvm.sourceTable = m.cctvm.sourceTable.Focused(false)
				m.cctvm.targetTable = m.cctvm.targetTable.Focused(false)
				m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.Focused(true)
			}
		}
	}

	m.cctvm.targetTable, cmd = m.cctvm.targetTable.Update(msg)
	cmds = append(cmds, cmd)

	m.cctvm.sourceTable, cmd = m.cctvm.sourceTable.Update(msg)
	cmds = append(cmds, cmd)

	m.cctvm.copyTaskTable, cmd = m.cctvm.copyTaskTable.Update(msg)
	cmds = append(cmds, cmd)

	// Add Custom footers
	m.cctvm.sourceTable = m.cctvm.sourceTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d", m.cctvm.sourceTable.CurrentPage(), m.cctvm.sourceTable.MaxPages(), m.cctvm.sourceTable.TotalRows()),
	)

	m.cctvm.targetTable = m.cctvm.targetTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d", m.cctvm.targetTable.CurrentPage(), m.cctvm.targetTable.MaxPages(), m.cctvm.targetTable.TotalRows()),
	)

	m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nMaps Selected %d", m.cctvm.copyTaskTable.CurrentPage(), m.cctvm.copyTaskTable.MaxPages(), m.cctvm.copyTaskTable.TotalRows()),
	)

	return m, tea.Batch(cmds...)
}

// Views - Functions that renders the UI based on the data in the model.
// https://github.com/charmbracelet/bubbletea/tree/master?tab=readme-ov-file#the-view-method

// The error view
func errorView(m model) string {
	tpl := "\nA fatal error occured while %s\n"
	tpl += "%s\n%s\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	return fmt.Sprintf(tpl, m.fatalError.context, errorImage, "Error Message: "+keywordStyle.Render(m.fatalError.text))
}

// The orchestrator view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.quitting {
		return "\n  See you next time!\n\n"
	}
	if m.fatalError != nil {
		return errorView(m)
	} else if !m.dscvm.sourceDatabaseChosen {
		s = databaseSourceChoicesView(m)
	} else if !m.dtcvm.targetDatabaseChosen {
		s = databaseTargetChoicesView(m)
	} else if !m.cctvm.collectionsChosen {
		s = collectionChoiceTableView(m)
	} else {
		s = copyStatusView(m)
	}
	return mainStyle.Render("\n" + s + "\n\n")
}

// The first view where user is chosing a source database
func databaseSourceChoicesView(m model) string {
	tpl := bannerStyle.Render(banner) + "\n"
	tpl += "Choose the " + keywordStyle.Render("source") + " database\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.dscvm.sourceDatabases {
		choices += fmt.Sprintf("%s\n", checkbox(choice, m.dscvm.sourceDatabaseChoice == i))
	}

	return fmt.Sprintf(tpl, choices)
}

// The second view where user is chosing a target database
func databaseTargetChoicesView(m model) string {
	tpl := bannerStyle.Render(banner) + "\n"
	tpl += "Choose the " + keywordStyle.Render("target") + " database\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.dtcvm.targetDatabases {
		choices += fmt.Sprintf("%s\n", checkbox(choice, m.dtcvm.targetDatabaseChoice == i))
	}

	return fmt.Sprintf(tpl, choices)
}

// The third view where use is choosing source and target collections
func collectionChoiceTableView(m model) string {
	tpl := bannerStyle.Render(banner) + "\n"
	tpl += "Map where the data lives and what collection it should be transfered to. You can choose one or more mappings.\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: change selection") + "\n" +
		subtleStyle.Render("left/right: change page") + "\n" +
		subtleStyle.Render("space: choose collection") + "\n" +
		subtleStyle.Render("enter: edit copy map") + "\n" +
		subtleStyle.Render("del: remove mapping (Copy table only)") + "\n"
	subtleStyle.Render("q, esc: quit")

	//body := strings.Builder{}

	// body.WriteString("Table demo with pagination! Press left/right to move pages, or use page up/down, or 'r' to jump to a random page\nPress 'a' for left table, 'b' for right table\nPress 'z' to reduce rows by 10, 'y' to increase rows by 10\nPress 'u' to decrease page size by 1, 'i' to increase page size by 1\nPress q or ctrl+c to quit\n\n")
	// body.WriteString("left/right: move pages")
	// body.WriteString("up/down: page up/down ")
	pad := lipgloss.NewStyle().Padding(1)

	tables := []string{
		lipgloss.JoinVertical(lipgloss.Center, "Choose source", pad.Render(m.cctvm.sourceTable.View())),
		lipgloss.JoinVertical(lipgloss.Center, "Choose target", pad.Render(m.cctvm.targetTable.View())),
		lipgloss.JoinVertical(lipgloss.Center, "Collection Copy Map", pad.Render(m.cctvm.copyTaskTable.View())),
	}

	var t = lipgloss.JoinHorizontal(lipgloss.Top, tables...)

	return fmt.Sprintf(tpl, t)
}

// The final view showing the status of the chosen copy tasks
func copyStatusView(m model) string {
	var progress, label string
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	tpl := bannerStyle.Render(banner) + "\n"
	tpl += fmt.Sprintf("Copying data to target database (%s)\n", keywordStyle.Render(m.dtcvm.targetDatabases[m.dtcvm.targetDatabaseChoice]))
	tpl += "%s\n\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	for i := 0; i < len(m.cctvm.copyTasks); i++ {

		label = "Copying..."
		spinner := fmt.Sprintf("%s %s", m.cctvm.copyTasks[i].spinner.View(), textStyle(label))
		if m.cctvm.copyTasks[i].complete {
			label = "Done"
			spinner = fmt.Sprintf("%s %s", "✅", textStyle(label))
		}
		progress += "\n\n" + keywordStyle.Render(m.cctvm.copyTasks[i].source+" -> "+m.cctvm.copyTasks[i].target) + " - " + spinner
	}

	return fmt.Sprintf(tpl, progress)
}

// Components

func checkbox(label string, checked bool) string {
	if checked {
		return checkboxStyle.Render("[x] " + label)
	}
	return fmt.Sprintf("[ ] %s", label)
}

// Utils

// Remove item from slice
func removeItem[T any](s []T, id int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:id]...)
	return append(ret, s[id+1:]...)
}

// Build an empty table using row data
func buildRows(tableData []table.RowData) []table.Row {
	rows := []table.Row{}

	for row := 0; row < len(tableData); row++ {
		rows = append(rows, table.NewRow(tableData[row]))
	}

	return rows
}

// Build an empty table using given columns headers
func buildTable(columns ...string) table.Model {
	c := []table.Column{}

	for i := 0; i < len(columns); i++ {
		columnName := columns[i]
		c = append(c, table.NewColumn(columnName, columnName, 20))
	}

	rows := buildRows([]table.RowData{})

	return table.New(c).
		WithRows(rows).
		HighlightStyle(checkboxStyle).
		HeaderStyle(lipgloss.NewStyle().Bold(true)).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#faa")),
			Data:  "-",
		})
}

// Build rows for both source and target collection tables
func (m *model) buildCollectionTableRows() {
	targetTableData := []table.RowData{}

	for i := 0; i < len(m.dtcvm.targetCollections); i++ {
		rowData := map[string]interface{}{targetColumnName: m.dtcvm.targetCollections[i]}
		targetTableData = append(targetTableData, rowData)
	}

	sourceTableData := []table.RowData{}

	for i := 0; i < len(m.dscvm.sourceCollections); i++ {
		rowData := map[string]interface{}{sourceColumnName: m.dscvm.sourceCollections[i]}
		sourceTableData = append(sourceTableData, rowData)
	}

	m.cctvm.targetTable = m.cctvm.targetTable.WithRows(buildRows(targetTableData))
	m.cctvm.sourceTable = m.cctvm.sourceTable.WithRows(buildRows(sourceTableData))
}

// Build rows for copyTask table
func (m *model) buildCollectionMapRows() {
	tableData := []table.RowData{}

	for i := 0; i < len(m.cctvm.copyTasks); i++ {
		rowData := map[string]interface{}{sourceColumnName: m.cctvm.copyTasks[i].source, targetColumnName: m.cctvm.copyTasks[i].target}
		tableData = append(tableData, rowData)
	}

	m.cctvm.copyTaskTable = m.cctvm.copyTaskTable.WithRows(buildRows(tableData))
}

// AddRow adds a new row to the table
func (t *TableData) AddRow(row map[string]interface{}) {
	t.Rows = append(t.Rows, row)
}
