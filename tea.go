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

type FatalError struct {
	Text    string
	Context string
}

type model struct {
	TargetDatabases         []string // Databases on server
	SourceDatabases         []string // Databases on server
	TargetDatabaseChoice    int      // Database chosen by user
	SourceDatabaseChoice    int      // Database chosen by user
	SourceDatabaseChosen    bool     // Has user made database selection
	TargetDatabaseChosen    bool     // Has user made database selection
	SourceCollections       []string // Collections in database
	TargetCollections       []string // Collections in database
	TargetCurrentCollection int      // Collection cursor is current on
	SourceCurrentCollection int      // Collection cursor is current on
	//CollectionChoices        []collectionChoice // Collections user has selected
	CollectionsChosen        bool        // Has user made database selection
	Quitting                 bool        // Has user quit application
	Storage                  storage     // Storage
	FatalError               *FatalError // Fatal Error details
	SourceTable              table.Model
	TargetTable              table.Model
	CopyTaskTable            table.Model
	RowCount                 int
	CopyTasks                []collectionCopyTask
	CurrentCopyTask          collectionCopyTask
	CollectionViewTableIndex int // What table is currently selected
}

// For messages that contain errors
func (e errMsg) Error() string { return e.err.Error() }

// Commands

func (m model) getSourceDatabases() tea.Msg {
	databases, err := m.Storage.getSourceDatabases()
	if err != nil {
		return errMsg{err, "getting source databases"}
	}

	return getSourceDatabasesMsg(databases)
}

func (m model) getTargetDatabases() tea.Msg {
	databases, err := m.Storage.getTargetDatabases()
	if err != nil {
		return errMsg{err, "getting target databases"}
	}

	return getTargetDatabasesMsg(databases)
}

func (m model) getCollections() tea.Msg {
	var collections collections
	var err error

	collections.target, err = m.Storage.getTargetCollections(m.TargetDatabases[m.TargetDatabaseChoice])
	if err != nil {
		return errMsg{err, "getting target collections"}
	}

	collections.source, err = m.Storage.getSourceCollections(m.SourceDatabases[m.SourceDatabaseChoice])
	if err != nil {
		return errMsg{err, "getting source collections"}
	}

	return getCollectionsMsg(collections)
}

func (m model) doCopy() []tea.Cmd {
	var err error
	var cmds []tea.Cmd

	for _, c := range m.CopyTasks {
		cmd := func() tea.Msg {
			err = m.Storage.copy(c.source, c.target)
			if err != nil {
				return errMsg{err, "copying records"}
			}

			return copyMsg{collectionId: c.id}
		}

		cmds = append(cmds, cmd)
	}

	return cmds
}

func (m model) Init() tea.Cmd {
	return m.getSourceDatabases
}

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Make sure these keys always quit
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	case getSourceDatabasesMsg:
		m.SourceDatabases = msg
		return m, tea.ClearScreen
	case getTargetDatabasesMsg:
		m.TargetDatabases = msg
		return m, tea.ClearScreen
	case getCollectionsMsg:
		m.SourceCollections = msg.source
		m.TargetCollections = msg.target
		m.regenTableRows()
		return m, tea.ClearScreen
	case copyMsg:
		for i := 0; i < len(m.CopyTasks); i++ {
			if msg.collectionId == m.CopyTasks[i].id {
				m.CopyTasks[i].complete = true
			}
		}
		return m, nil
	case spinner.TickMsg:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)
		for i := 0; i < len(m.CopyTasks); i++ {
			m.CopyTasks[i].spinner, cmd = m.CopyTasks[i].spinner.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		return m, tea.Batch(cmds...)
	case errMsg:
		m.FatalError = &FatalError{Text: msg.err.Error(), Context: msg.context}

		return m, tea.ClearScreen
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.SourceDatabaseChosen {
		return updateSourceDatabaseChoices(msg, m)
	} else if !m.TargetDatabaseChosen {
		return updateTargetDatabaseChoices(msg, m)
	} else if !m.CollectionsChosen {
		return updateTable(msg, m)
		//return updateCollectionChoices(msg, m)
	}

	return m, nil
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return "\n  See you next time!\n\n"
	}
	if m.FatalError != nil {
		return errorView(m)
	} else if !m.SourceDatabaseChosen {
		s = databaseSourceChoicesView(m)
	} else if !m.TargetDatabaseChosen {
		s = databaseTargetChoicesView(m)
	} else if !m.CollectionsChosen {
		s = collectionChoiceTableView(m)
	} else {
		s = copyView(m)
	}
	return mainStyle.Render("\n" + s + "\n\n")
}

// Update functions

// Update loop for the first view where you're choosing a database.
func updateSourceDatabaseChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.SourceDatabaseChoice++
			if m.SourceDatabaseChoice >= len(m.SourceDatabases) {
				m.SourceDatabaseChoice = len(m.SourceDatabases) - 1
			}
		case "up":
			m.SourceDatabaseChoice--
			if m.SourceDatabaseChoice < 0 {
				m.SourceDatabaseChoice = 0
			}
		case "enter":
			m.SourceDatabaseChosen = true
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
			m.TargetDatabaseChoice++
			if m.TargetDatabaseChoice >= len(m.TargetDatabases) {
				m.TargetDatabaseChoice = len(m.TargetDatabases) - 1
			}
		case "up":
			m.TargetDatabaseChoice--
			if m.TargetDatabaseChoice < 0 {
				m.TargetDatabaseChoice = 0
			}
		case "enter":
			m.TargetDatabaseChosen = true
			return m, m.getCollections
		}
	}

	return m, nil
}

// Update loop for the first view where you're choosing collections.
// func updateCollectionChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "down":
// 			m.CurrentCollection++
// 			if m.CurrentCollection >= len(m.SourceCollections) {
// 				m.CurrentCollection = len(m.SourceCollections) - 1
// 			}
// 		case "up":
// 			m.CurrentCollection--
// 			if m.CurrentCollection < 0 {
// 				m.CurrentCollection = 0
// 			}
// 		case " ":
// 			if !containsChoice(m.CollectionChoices, m.CurrentCollection) {
// 				var s = spinner.New()
// 				s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
// 				s.Spinner = spinner.Line

// 				collectionChoice := collectionChoice{
// 					Id:      m.CurrentCollection,
// 					Name:    m.SourceCollections[m.CurrentCollection],
// 					Spinner: s}
// 				m.CollectionChoices = append(m.CollectionChoices, collectionChoice)
// 			} else {
// 				for i := 0; i < len(m.CollectionChoices); i++ {
// 					if m.CollectionChoices[i].Id == m.CurrentCollection {
// 						m.CollectionChoices = RemoveChoice(m.CollectionChoices, i)
// 					}
// 				}
// 			}
// 			return m, nil
// 		case "enter":

// 			if len(m.CollectionChoices) == 0 {
// 				return m, nil
// 			}

// 			m.CollectionsChosen = true

// 			var cmds []tea.Cmd
// 			var c = m.doCopy()
// 			cmds = append(cmds, c...)

// 			for i := 0; i < len(m.CollectionChoices); i++ {
// 				cmd := func() tea.Msg {
// 					return m.CollectionChoices[i].Spinner.Tick()
// 				}

// 				cmds = append(cmds, cmd)
// 			}

// 			return m, tea.Batch(cmds...)
// 		}
// 	}

// 	return m, nil
// }

func updateTable(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
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
			m.SourceTable = m.SourceTable.Focused(true)
			m.TargetTable = m.TargetTable.Focused(false)
			m.CopyTaskTable = m.CopyTaskTable.Focused(false)

		case "b":
			m.SourceTable = m.SourceTable.Focused(false)
			m.TargetTable = m.TargetTable.Focused(true)
			m.CopyTaskTable = m.CopyTaskTable.Focused(false)

		case "enter":
			if len(m.CopyTasks) != 0 {
				m.SourceTable = m.SourceTable.Focused(false)
				m.TargetTable = m.TargetTable.Focused(false)
				m.CopyTaskTable = m.CopyTaskTable.Focused(true)
			}

		case "s":
			if len(m.CopyTasks) == 0 {
				return m, nil
			}

			m.CollectionsChosen = true

			var cmds []tea.Cmd
			var c = m.doCopy()
			cmds = append(cmds, c...)

			for i := 0; i < len(m.CopyTasks); i++ {
				cmd := func() tea.Msg {
					return m.CopyTasks[i].spinner.Tick()
				}

				cmds = append(cmds, cmd)
			}

			return m, tea.Batch(cmds...)

		case "u":
			m.SourceTable = m.SourceTable.WithPageSize(m.SourceTable.PageSize() - 1)
			m.TargetTable = m.TargetTable.WithPageSize(m.TargetTable.PageSize() - 1)

		case "i":
			m.SourceTable = m.SourceTable.WithPageSize(m.SourceTable.PageSize() + 1)
			m.TargetTable = m.TargetTable.WithPageSize(m.TargetTable.PageSize() + 1)

		case "r":
			m.SourceTable = m.SourceTable.WithCurrentPage(rand.Intn(m.SourceTable.MaxPages()) + 1)
			m.TargetTable = m.TargetTable.WithCurrentPage(rand.Intn(m.TargetTable.MaxPages()) + 1)

		case "z":
			if m.RowCount < 10 {
				break
			}

			m.RowCount -= 10
			m.regenTableRows()

		case "x":
			m.RowCount += 10
			m.regenTableRows()

		case "delete":
			// Only delete form copy map table
			if m.CollectionViewTableIndex == 2 {
				//Todo Delete mappings
			}

		case " ":
			if m.SourceTable.GetFocused() {
				m.CollectionViewTableIndex++
				row := m.SourceTable.HighlightedRow()

				var value = row.Data[sourceColumnName].(string)
				m.CurrentCopyTask.source = value

				// Delete collection so it can't be selected again
				var i = m.SourceTable.GetHighlightedRowIndex()
				m.SourceCollections = RemoveItem(m.SourceCollections, i)
				m.regenTableRows()
				m.SourceTable = m.SourceTable.Focused(false)
				m.TargetTable = m.TargetTable.Focused(true)

			} else if m.TargetTable.GetFocused() {
				m.CollectionViewTableIndex++
				row := m.TargetTable.HighlightedRow()
				var value = row.Data[targetColumnName].(string)
				m.CurrentCopyTask.target = value

				// Delete collection so it can't be selected again
				var i = m.TargetTable.GetHighlightedRowIndex()
				m.TargetCollections = RemoveItem(m.TargetCollections, i)
				m.regenTableRows()
				m.SourceTable = m.SourceTable.Focused(true)
				m.TargetTable = m.TargetTable.Focused(false)
			}

			// User has chose target and source, so add to copy map
			if m.CollectionViewTableIndex == 2 {
				m.CopyTasks = append(m.CopyTasks, m.CurrentCopyTask)
				m.regenCollectionMapRows()
				m.CollectionViewTableIndex = 0
			}

			// No more viable copy maps to be selected
			if len(m.SourceCollections) == 0 && len(m.TargetCollections) == 0 {
				m.SourceTable = m.SourceTable.Focused(false)
				m.TargetTable = m.TargetTable.Focused(false)
				m.CopyTaskTable = m.CopyTaskTable.Focused(true)
			}
		}
	}

	m.TargetTable, cmd = m.TargetTable.Update(msg)
	cmds = append(cmds, cmd)

	m.SourceTable, cmd = m.SourceTable.Update(msg)
	cmds = append(cmds, cmd)

	m.CopyTaskTable, cmd = m.CopyTaskTable.Update(msg)
	cmds = append(cmds, cmd)

	// Add Custom footers
	m.SourceTable = m.SourceTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d", m.SourceTable.CurrentPage(), m.SourceTable.MaxPages(), m.SourceTable.TotalRows()),
	)

	m.TargetTable = m.TargetTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d", m.TargetTable.CurrentPage(), m.TargetTable.MaxPages(), m.TargetTable.TotalRows()),
	)

	m.CopyTaskTable = m.CopyTaskTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nMaps Selected %d", m.CopyTaskTable.CurrentPage(), m.CopyTaskTable.MaxPages(), m.CopyTaskTable.TotalRows()),
	)

	return m, tea.Batch(cmds...)
}

// Views

// The error view
func errorView(m model) string {
	tpl := "\nA fatal error occured while %s\n"
	tpl += "%s\n%s\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	return fmt.Sprintf(tpl, m.FatalError.Context, errorImage, "Error Message: "+keywordStyle.Render(m.FatalError.Text))
}

// The first view, where you're choosing a source database
func databaseSourceChoicesView(m model) string {
	tpl := banner + "\n"
	tpl += "Choose the source database\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.SourceDatabases {
		choices += fmt.Sprintf("%s\n", checkbox(choice, m.SourceDatabaseChoice == i))
	}

	return fmt.Sprintf(tpl, choices)
}

func databaseTargetChoicesView(m model) string {
	tpl := banner + "\n"
	tpl += "Choose the target database\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.TargetDatabases {
		choices += fmt.Sprintf("%s\n", checkbox(choice, m.TargetDatabaseChoice == i))
	}

	return fmt.Sprintf(tpl, choices)
}

// The second view, where you're choosing a collections
func collectionChoiceTableView(m model) string {
	tpl := banner + "\n"
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
		lipgloss.JoinVertical(lipgloss.Center, "Choose source", pad.Render(m.SourceTable.View())),
		lipgloss.JoinVertical(lipgloss.Center, "Choose target", pad.Render(m.TargetTable.View())),
		lipgloss.JoinVertical(lipgloss.Center, "Collection Copy Map", pad.Render(m.CopyTaskTable.View())),
	}

	var t = lipgloss.JoinHorizontal(lipgloss.Top, tables...)

	return fmt.Sprintf(tpl, t)
}

// The copy view shown after a collections has been chosen
func copyView(m model) string {
	var progress, label string
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	tpl := banner + "\n"
	tpl += fmt.Sprintf("Copying data to target database (%s)\n", keywordStyle.Render(m.TargetDatabases[m.TargetDatabaseChoice]))
	tpl += "%s\n\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	for i := 0; i < len(m.CopyTasks); i++ {

		label = "Copying..."
		spinner := fmt.Sprintf("%s %s", m.CopyTasks[i].spinner.View(), textStyle(label))
		if m.CopyTasks[i].complete {
			label = "Done"
			spinner = fmt.Sprintf("%s %s", "✅", textStyle(label))
		}
		progress += "\n\n" + keywordStyle.Render(m.CopyTasks[i].source+" -> "+m.CopyTasks[i].target) + " - " + spinner
	}

	return fmt.Sprintf(tpl, progress)
}

// Checkbox used when user can select only one at a time
func checkbox(label string, checked bool) string {
	if checked {
		return checkboxStyle.Render("[x] " + label)
	}
	return fmt.Sprintf("[ ] %s", label)
}

// Utils

// Remove item from slice.
func RemoveItem[T any](s []T, id int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:id]...)
	return append(ret, s[id+1:]...)
}

func genRows(collections []string, columnName string) []table.Row {
	rows := []table.Row{}

	for row := 0; row < len(collections); row++ {
		rowData := table.RowData{}

		for column := 0; column < 1; column++ {
			columnStr := columnName
			rowData[columnStr] = collections[row]
		}

		rows = append(rows, table.NewRow(rowData))
	}

	return rows
}

func genRowsv2(tableData []table.RowData) []table.Row {
	rows := []table.Row{}

	for row := 0; row < len(tableData); row++ {
		rows = append(rows, table.NewRow(tableData[row]))
	}

	return rows
}

func genTable(columns ...string) table.Model {
	c := []table.Column{}

	for i := 0; i < len(columns); i++ {
		columnName := columns[i]
		c = append(c, table.NewColumn(columnName, columnName, 20))
	}

	rows := genRowsv2([]table.RowData{})

	return table.New(c).
		WithRows(rows).
		HighlightStyle(checkboxStyle).
		HeaderStyle(lipgloss.NewStyle().Bold(true)).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#faa")),
			Data:  "-",
		})
}

func (m *model) regenTableRows() {
	m.TargetTable = m.TargetTable.WithRows(genRows(m.TargetCollections, "Target Collections"))
	m.SourceTable = m.SourceTable.WithRows(genRows(m.SourceCollections, "Source Collections"))
}

func (m *model) regenCollectionMapRows() {

	tableData := []table.RowData{}

	for i := 0; i < len(m.CopyTasks); i++ {
		rowData := map[string]interface{}{sourceColumnName: m.CopyTasks[i].source, targetColumnName: m.CopyTasks[i].target}
		tableData = append(tableData, rowData)
	}

	m.CopyTaskTable = m.CopyTaskTable.WithRows(genRowsv2(tableData))
}

type TableData struct {
	Columns []string
	Rows    []map[string]interface{}
}

// AddRow adds a new row to the table
func (t *TableData) AddRow(row map[string]interface{}) {
	t.Rows = append(t.Rows, row)
}
