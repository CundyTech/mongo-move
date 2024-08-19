package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	targetCollectionsColumnName = "Target Collections"
	sourceCollectionsColumnName = "Source Collections"
	targetDatabasesColumnName   = "Target Databases"
	sourceDatabasesColumnName   = "Source Databases"
	taskMapColumnName           = "Collections Map"
	recordsCountColumnName      = "Records"
	CopyStatusColumnName        = "Copy Status"
	progressBarWidth            = 71
	dotChar                     = " • "
	banner                      = `
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
	green         = lipgloss.NewStyle().Foreground(lipgloss.Color("#54ad48"))
	keywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
)

type copyMsg struct {
	collectionId int
}

type collection struct {
	name  string
	count int64
}

type collections struct {
	target []collection
	source []collection
}

type databases struct {
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
	getDatabasesMsg      databases
	databasesLoadedMsg   bool
	getCollectionsMsg    collections
	collectionsLoadedMsg bool
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
type collectionChoicesViewModel struct {
	sourceTable         table.Model // Table that displays collections in the source database
	sourceTableFiltered bool
	targetTable         table.Model // Table that displays collections in the target database
	targetTableFiltered bool
	copyTaskTable       table.Model          // Table that displays chosen source and target collection maps
	copyTasks           []collectionCopyTask // Collection of source and targets where data will be moved from and to respectively
	currentCopyTask     collectionCopyTask   // Vurrent user selection of source and target collections
	currentTableIndex   int                  // Index of the table is currently in use by user. 0 = sourceTable, 1 = targetTable and 2 = copyTaskTable
	pageSize            int                  // Default size of a page of all tables
	rowCount            int                  // The amount of rows in a table
	collectionsChosen   bool                 // Has user made collection choices
	collectionsLoaded   bool
	debounce            time.Duration // debounce duraiton for loading spinner
	altscreen           bool
}

type databaseChoicesViewModel struct {
	sourceDatabases         []string // Databases on server
	sourceDatabaseChoice    string   // Database chosen by user
	databasesChosen         bool     // Has user made database selections
	databasesLoaded         bool
	sourceCollections       []collection // Collections in database
	sourceCurrentCollection int          // Collection cursor is current on
	sourcePageSize          int          // Default size of a page of all tables
	sourceRowCount          int          // The amount of rows in a table
	sourceTable             table.Model  // Table that displays collections in the source database
	sourceTableFiltered     bool
	targetDatabases         []string     // Databases on server
	targetDatabaseChoice    string       // Database chosen by user
	targetCollections       []collection // Collections in database
	targetCurrentCollection int          // Collection cursor is current on
	targetPageSize          int          // Default size of a page of all tables
	targetRowCount          int          // The amount of rows in a table
	targetTable             table.Model  // Table that displays collections in the target database
	targetTableFiltered     bool
	currentTableIndex       int           // Index of the table is currently in use by user. 0 = sourceTable, 1 = targetTable
	debounce                time.Duration // debounce duraiton for loading spinner
}

// Main model
type model struct {
	keyBindings       keyModel
	storage           storage                    // Storage
	fatalError        *fatalError                // Fatal Error details
	databaseChoices   databaseChoicesViewModel   // Model for databaseChoicesView view
	collectionChoices collectionChoicesViewModel // Model for collectionChoicesTable view
	spinner           spinner.Model
}

// Init function that returns an initial command for the application to run
func (m model) Init() tea.Cmd {

	var cmds []tea.Cmd
	cmds = append(cmds, m.spinner.Tick)
	cmds = append(cmds, m.getSourceDatabases)

	return tea.Batch(cmds...)
}

// Commands -  Functions that perform some I/O and then return a Msg.
// https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands/

func (m model) getSourceDatabases() tea.Msg {
	var databases databases
	var err error

	databases.source, err = m.storage.getSourceDatabases()
	if err != nil {
		return errMsg{err, "getting source databases"}
	}

	databases.target, err = m.storage.getTargetDatabases()
	if err != nil {
		return errMsg{err, "getting target databases"}
	}

	return getDatabasesMsg(databases)
}

func (m model) getCollections() tea.Msg {
	var collections collections
	var err error

	collections.target, err = m.storage.getTargetCollections(m.databaseChoices.targetDatabaseChoice)
	if err != nil {
		return errMsg{err, "getting target collections"}
	}

	collections.source, err = m.storage.getSourceCollections(m.databaseChoices.sourceDatabaseChoice)
	if err != nil {
		return errMsg{err, "getting source collections"}
	}

	return getCollectionsMsg(collections)
}

func (m model) copyData() []tea.Cmd {
	var err error
	var cmds []tea.Cmd

	for _, c := range m.collectionChoices.copyTasks {
		cmd := func() tea.Msg {
			err = m.storage.copy(c.source, c.target, m.databaseChoices.sourceDatabaseChoice, m.databaseChoices.targetDatabaseChoice)
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
		if k == "q" || k == "ctrl+c" {
			m.keyBindings.quitting = true
			return m, tea.Quit
		}
	case getDatabasesMsg:
		m.databaseChoices.sourceDatabases = msg.source
		m.databaseChoices.targetDatabases = msg.target
		m.buildSourceDatabaseTableRows()
		m.buildTargetDatabaseTableRows()

		// Debounce spinner
		return m, tea.Tick(time.Duration(m.databaseChoices.debounce), func(_ time.Time) tea.Msg {
			return databasesLoadedMsg(true)
		})
	case databasesLoadedMsg:
		m.databaseChoices.databasesLoaded = true
		return m, tea.ClearScreen
	case getCollectionsMsg:
		m.databaseChoices.sourceCollections = msg.source
		m.databaseChoices.targetCollections = msg.target
		m.buildCollectionTableRows()

		// Debounce spinner
		return m, tea.Tick(time.Duration(m.databaseChoices.debounce), func(_ time.Time) tea.Msg {
			return collectionsLoadedMsg(true)
		})

	case collectionsLoadedMsg:
		m.collectionChoices.collectionsLoaded = true
		return m, tea.ClearScreen
	case copyMsg:
		for i := 0; i < len(m.collectionChoices.copyTasks); i++ {
			if msg.collectionId == m.collectionChoices.copyTasks[i].id {
				m.collectionChoices.copyTasks[i].complete = true
			}
		}
		return m, nil
	case spinner.TickMsg:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)

		for i := 0; i < len(m.collectionChoices.copyTasks); i++ {
			m.collectionChoices.copyTasks[i].spinner, cmd = m.collectionChoices.copyTasks[i].spinner.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	case errMsg:
		m.fatalError = &fatalError{text: msg.err.Error(), context: msg.context}

		return m, tea.ClearScreen
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !(m.databaseChoices.databasesChosen) {
		return updateDatabaseChoices(msg, m)
	} else if !m.collectionChoices.collectionsChosen {
		return updateCollectionChoiceTable(msg, m)
	}

	return m, nil
}

// Update loop for the first view where you're choosing a database.
func updateDatabaseChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyBindings.keys.QuitFilter):
			if m.databaseChoices.sourceTable.GetFocused() {
				m.databaseChoices.sourceTableFiltered = false
			} else if m.databaseChoices.targetTable.GetFocused() {
				m.databaseChoices.targetTableFiltered = false
			}
		case key.Matches(msg, m.keyBindings.keys.Filter):
			// Set as as filtered so we can update the UI to give a clue before user input
			if m.databaseChoices.sourceTable.GetFocused() {
				m.databaseChoices.sourceTableFiltered = true
			} else if m.databaseChoices.targetTable.GetFocused() {
				m.databaseChoices.targetTableFiltered = true
			}
		case key.Matches(msg, m.keyBindings.keys.Quit):
			m.keyBindings.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keyBindings.keys.Select):
			if m.databaseChoices.sourceTable.GetFocused() {
				m.databaseChoices.currentTableIndex++
				row := m.databaseChoices.sourceTable.HighlightedRow()

				var value = row.Data[sourceDatabasesColumnName].(string)
				m.databaseChoices.sourceDatabaseChoice = value

				m.databaseChoices.sourceTable = m.databaseChoices.sourceTable.Focused(false)
				m.databaseChoices.targetTable = m.databaseChoices.targetTable.Focused(true)

			} else if m.databaseChoices.targetTable.GetFocused() {
				m.databaseChoices.currentTableIndex++
				row := m.databaseChoices.targetTable.HighlightedRow()
				var value = row.Data[targetDatabasesColumnName].(string)
				m.databaseChoices.targetDatabaseChoice = value

				m.databaseChoices.databasesChosen = true

				return m, m.getCollections
			}
		}

	}

	m.databaseChoices.targetTable, cmd = m.databaseChoices.targetTable.Update(msg)
	cmds = append(cmds, cmd)

	m.databaseChoices.sourceTable, cmd = m.databaseChoices.sourceTable.Update(msg)
	cmds = append(cmds, cmd)

	var stfilterText string
	if len(m.databaseChoices.sourceTable.GetCurrentFilter()) > 0 || m.databaseChoices.sourceTableFiltered {
		stfilterText = fmt.Sprintf("\nFilter: %s", m.databaseChoices.sourceTable.GetCurrentFilter())
	}

	// Add Custom footers
	m.databaseChoices.sourceTable = m.databaseChoices.sourceTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d \n%s",
			m.databaseChoices.sourceTable.CurrentPage(),
			m.databaseChoices.sourceTable.MaxPages(),
			m.databaseChoices.sourceTable.TotalRows(),
			stfilterText),
	)

	var ttfilterText string
	if len(m.databaseChoices.targetTable.GetCurrentFilter()) > 0 || m.databaseChoices.targetTableFiltered {
		ttfilterText = fmt.Sprintf("\nFilter: %s", m.databaseChoices.targetTable.GetCurrentFilter())
	} else {
		ttfilterText = ""
	}

	m.databaseChoices.targetTable = m.databaseChoices.targetTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d \n%s",
			m.databaseChoices.targetTable.CurrentPage(),
			m.databaseChoices.targetTable.MaxPages(),
			m.databaseChoices.targetTable.TotalRows(),
			ttfilterText,
		),
	)

	return m, tea.Batch(cmds...)
}

func updateCollectionChoiceTable(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(true)
			m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(false)
			m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.Focused(false)

		case "b":
			m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(false)
			m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(true)
			m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.Focused(false)

		case "enter":
			if len(m.collectionChoices.copyTasks) != 0 {
				m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(false)
				m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(false)
				m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.Focused(true)
			}
		case "tab":
			var cmd tea.Cmd
			if m.collectionChoices.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}
			m.collectionChoices.altscreen = !m.collectionChoices.altscreen
			return m, cmd

		case "s":
			if len(m.collectionChoices.copyTasks) == 0 {
				return m, nil
			}

			m.collectionChoices.collectionsChosen = true

			var cmds []tea.Cmd
			var c = m.copyData()
			cmds = append(cmds, c...)

			for i := 0; i < len(m.collectionChoices.copyTasks); i++ {
				cmd := func() tea.Msg {
					return m.collectionChoices.copyTasks[i].spinner.Tick()
				}

				cmds = append(cmds, cmd)
			}

			return m, tea.Batch(cmds...)

		case "u":
			m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.WithPageSize(m.collectionChoices.sourceTable.PageSize() - 1)
			m.collectionChoices.targetTable = m.collectionChoices.targetTable.WithPageSize(m.collectionChoices.targetTable.PageSize() - 1)

		case "i":
			m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.WithPageSize(m.collectionChoices.sourceTable.PageSize() + 1)
			m.collectionChoices.targetTable = m.collectionChoices.targetTable.WithPageSize(m.collectionChoices.targetTable.PageSize() + 1)

		case "r":
			m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.WithCurrentPage(rand.Intn(m.collectionChoices.sourceTable.MaxPages()) + 1)
			m.collectionChoices.targetTable = m.collectionChoices.targetTable.WithCurrentPage(rand.Intn(m.collectionChoices.targetTable.MaxPages()) + 1)

		case "z":
			if m.collectionChoices.rowCount < 10 {
				break
			}

			m.collectionChoices.rowCount -= 10
			m.buildCollectionTableRows()

		case "x":
			m.collectionChoices.rowCount += 10
			m.buildCollectionTableRows()

		case "delete":
			// Only delete form copy map table
			if m.collectionChoices.currentTableIndex == 2 {
				//Todo Delete mappings
			}

		case " ":
			if m.collectionChoices.sourceTable.GetFocused() {
				m.collectionChoices.currentTableIndex++
				row := m.collectionChoices.sourceTable.HighlightedRow()

				var value = row.Data[sourceCollectionsColumnName].(string)
				m.collectionChoices.currentCopyTask.source = value

				// Delete collection so it can't be selected again
				var i = m.collectionChoices.sourceTable.GetHighlightedRowIndex()
				m.databaseChoices.sourceCollections = removeItem(m.databaseChoices.sourceCollections, i)
				m.buildCollectionTableRows()
				m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(false)
				m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(true)

			} else if m.collectionChoices.targetTable.GetFocused() {
				m.collectionChoices.currentTableIndex++
				row := m.collectionChoices.targetTable.HighlightedRow()
				var value = row.Data[targetCollectionsColumnName].(string)
				m.collectionChoices.currentCopyTask.target = value

				// Delete collection so it can't be selected again
				var i = m.collectionChoices.targetTable.GetHighlightedRowIndex()
				m.databaseChoices.targetCollections = removeItem(m.databaseChoices.targetCollections, i)
				m.buildCollectionTableRows()
				m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(true)
				m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(false)
			}

			// User has chose target and source, so add to copy map
			if m.collectionChoices.currentTableIndex == 2 {
				m.collectionChoices.copyTasks = append(m.collectionChoices.copyTasks, m.collectionChoices.currentCopyTask)
				m.buildCollectionMapRows()
				m.collectionChoices.currentTableIndex = 0
			}

			// No more viable copy maps to be selected
			if len(m.databaseChoices.sourceCollections) == 0 && len(m.databaseChoices.targetCollections) == 0 {
				m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.Focused(false)
				m.collectionChoices.targetTable = m.collectionChoices.targetTable.Focused(false)
				m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.Focused(true)
			}

		case "/":
			// Set as as filtered so we can update the UI to give a clue before user input
			if m.collectionChoices.sourceTable.GetFocused() {
				m.collectionChoices.sourceTableFiltered = true
			} else if m.collectionChoices.targetTable.GetFocused() {
				m.collectionChoices.targetTableFiltered = true
			}
		case "esc":
			if m.collectionChoices.sourceTable.GetFocused() {
				m.collectionChoices.sourceTableFiltered = false
			} else if m.collectionChoices.targetTable.GetFocused() {
				m.collectionChoices.targetTableFiltered = false
			}
		}
	}

	m.collectionChoices.targetTable, cmd = m.collectionChoices.targetTable.Update(msg)
	cmds = append(cmds, cmd)

	m.collectionChoices.sourceTable, cmd = m.collectionChoices.sourceTable.Update(msg)
	cmds = append(cmds, cmd)

	m.collectionChoices.copyTaskTable, cmd = m.collectionChoices.copyTaskTable.Update(msg)
	cmds = append(cmds, cmd)

	// Add Custom footers

	var stfilterText string
	if len(m.collectionChoices.sourceTable.GetCurrentFilter()) > 0 || m.collectionChoices.sourceTableFiltered {
		stfilterText = fmt.Sprintf("\nFilter: %s", m.collectionChoices.sourceTable.GetCurrentFilter())
	}

	m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d \n%s",
			m.collectionChoices.sourceTable.CurrentPage(),
			m.collectionChoices.sourceTable.MaxPages(),
			m.collectionChoices.sourceTable.TotalRows(),
			stfilterText),
	)

	var ttfilterText string
	if len(m.collectionChoices.targetTable.GetCurrentFilter()) > 0 || m.collectionChoices.targetTableFiltered {
		ttfilterText = fmt.Sprintf("\nFilter: %s", m.collectionChoices.targetTable.GetCurrentFilter())
	}

	m.collectionChoices.targetTable = m.collectionChoices.targetTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nCollections %d \n%s",
			m.collectionChoices.targetTable.CurrentPage(),
			m.collectionChoices.targetTable.MaxPages(),
			m.collectionChoices.targetTable.TotalRows(),
			ttfilterText),
	)

	var cttfilterText string
	if len(m.collectionChoices.copyTaskTable.GetCurrentFilter()) > 0 {
		cttfilterText = fmt.Sprintf("\nFilter: %s", m.collectionChoices.copyTaskTable.GetCurrentFilter())
	}

	m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.WithStaticFooter(
		fmt.Sprintf("Page %d/%d \nMaps Selected %d \n%s",
			m.collectionChoices.copyTaskTable.CurrentPage(),
			m.collectionChoices.copyTaskTable.MaxPages(),
			m.collectionChoices.copyTaskTable.TotalRows(),
			cttfilterText),
	)

	return m, tea.Batch(cmds...)
}

// Views - Functions that renders the UI based on the data in the model.
// https://github.com/charmbracelet/bubbletea/tree/master?tab=readme-ov-file#the-view-method

// The error view
func errorView(m model) string {
	tpl := "\nA fatal error occured while %s\n"
	tpl += "%s\n%s\n\n"
	tpl += subtleStyle.Render("q: quit")

	return fmt.Sprintf(tpl, m.fatalError.context, errorImage, "Error Message: "+keywordStyle.Render(m.fatalError.text))
}

// The orchestrator view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.keyBindings.quitting {
		return "\n  See you next time!\n\n"
	}
	if m.fatalError != nil {
		return errorView(m)
	} else if !m.databaseChoices.databasesChosen {
		s = databaseChoicesView(m)
	} else if !m.collectionChoices.collectionsChosen {
		s = collectionChoiceTableView(m)
	} else {
		s = copyStatusView(m)
	}
	return mainStyle.Render("\n" + s + "\n\n")
}

// The first view where user is chosing a source database
func databaseChoicesView(m model) string {
	tpl := green.Render(banner) + "\n"
	tpl += "Choose the databases where the source and target collections reside \n\n"
	tpl += "%s"
	tpl += m.databaseChoicesHelp()

	var view string
	if !m.databaseChoices.databasesLoaded {
		spinner := fmt.Sprintf("\n %s%s\n\n", m.spinner.View(), " Fetching databases...")
		view = lipgloss.PlaceHorizontal(60, lipgloss.Center, spinner)
	} else {
		pad := lipgloss.NewStyle().Padding(1)
		tables := []string{
			lipgloss.JoinVertical(lipgloss.Center, pad.Render(m.databaseChoices.sourceTable.View())),
			lipgloss.JoinVertical(lipgloss.Center, pad.Render(m.databaseChoices.targetTable.View())),
		}
		view = lipgloss.JoinHorizontal(lipgloss.Top, tables...)
	}

	return fmt.Sprintf(tpl, view)
}

// The third view where use is choosing source and target collections
func collectionChoiceTableView(m model) string {
	tpl := green.Render(banner) + "\n"
	tpl += "%s\n%s"

	var view string
	var title string
	pad := lipgloss.NewStyle().Padding(1)

	if !m.collectionChoices.collectionsLoaded {
		title = "Select one more more pairs of source and target collections and we'll handle the move once you are finshed"
		spinner := fmt.Sprintf("\n %s%s\n\n", m.spinner.View(), " Fetching Collections...")
		view = lipgloss.PlaceHorizontal(80, lipgloss.Center, spinner)
		tpl += m.collectionChoicesHelp()
	} else {
		var tables []string
		if m.collectionChoices.altscreen {
			tpl += m.collectionChoicesCopyHelp()
			title = "Amend choices or press enter to start coping data"
			tables = []string{
				lipgloss.JoinVertical(lipgloss.Center, pad.Render(m.collectionChoices.copyTaskTable.View())),
			}
		} else {
			title = "Select one more more pairs of source and target collections and we'll handle the move once you are finshed"
			tables = []string{
				lipgloss.JoinVertical(lipgloss.Center, pad.Render(m.collectionChoices.sourceTable.View())),
				lipgloss.JoinVertical(lipgloss.Center, pad.Render(m.collectionChoices.targetTable.View())),
			}
			tpl += m.collectionChoicesHelp()
		}
		view = lipgloss.JoinHorizontal(lipgloss.Top, tables...)
	}

	return fmt.Sprintf(tpl, title, view)
}

// The final view showing the status of the chosen copy tasks
func copyStatusView(m model) string {
	var progress, label string
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	tpl := green.Render(banner) + "\n"
	tpl += fmt.Sprintf("Copying data to target database (%s)\n", keywordStyle.Render(m.databaseChoices.targetDatabaseChoice))
	tpl += "%s\n\n\n"

	for i := 0; i < len(m.collectionChoices.copyTasks); i++ {

		label = "Copying..."
		spinner := fmt.Sprintf("%s %s", m.collectionChoices.copyTasks[i].spinner.View(), textStyle(label))
		if m.collectionChoices.copyTasks[i].complete {
			label = "Done"
			spinner = fmt.Sprintf("%s %s", "✅", textStyle(label))
		}
		progress += "\n\n" + keywordStyle.Render(m.collectionChoices.copyTasks[i].source+" -> "+m.collectionChoices.copyTasks[i].target) + " - " + spinner
	}

	return fmt.Sprintf(tpl, progress)
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
		c = append(c, table.NewColumn(columnName, columnName, 25).
			WithFiltered(true))
	}

	rows := buildRows([]table.RowData{})

	return table.New(c).
		Filtered(true).
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

	for i := 0; i < len(m.databaseChoices.targetCollections); i++ {
		rowData := map[string]interface{}{
			targetCollectionsColumnName: m.databaseChoices.targetCollections[i].name,
			recordsCountColumnName:      m.databaseChoices.targetCollections[i].count}
		targetTableData = append(targetTableData, rowData)
	}

	sourceTableData := []table.RowData{}

	for i := 0; i < len(m.databaseChoices.sourceCollections); i++ {
		rowData := map[string]interface{}{
			sourceCollectionsColumnName: m.databaseChoices.sourceCollections[i].name,
			recordsCountColumnName:      m.databaseChoices.sourceCollections[i].count}
		sourceTableData = append(sourceTableData, rowData)
	}

	m.collectionChoices.targetTable = m.collectionChoices.targetTable.WithRows(buildRows(targetTableData))
	m.collectionChoices.sourceTable = m.collectionChoices.sourceTable.WithRows(buildRows(sourceTableData))
}

// Build rows for source database tables
func (m *model) buildSourceDatabaseTableRows() {
	sourceTableData := []table.RowData{}

	for i := 0; i < len(m.databaseChoices.sourceDatabases); i++ {
		rowData := map[string]interface{}{sourceDatabasesColumnName: m.databaseChoices.sourceDatabases[i]}
		sourceTableData = append(sourceTableData, rowData)
	}

	m.databaseChoices.sourceTable = m.databaseChoices.sourceTable.WithRows(buildRows(sourceTableData))
}

// Build rows for target database tables
func (m *model) buildTargetDatabaseTableRows() {
	targetTableData := []table.RowData{}

	for i := 0; i < len(m.databaseChoices.targetDatabases); i++ {
		rowData := map[string]interface{}{targetDatabasesColumnName: m.databaseChoices.targetDatabases[i]}
		targetTableData = append(targetTableData, rowData)
	}

	m.databaseChoices.targetTable = m.databaseChoices.targetTable.WithRows(buildRows(targetTableData))
}

// Build rows for copyTask table
func (m *model) buildCollectionMapRows() {
	tableData := []table.RowData{}

	for i := 0; i < len(m.collectionChoices.copyTasks); i++ {
		rowData := map[string]interface{}{sourceCollectionsColumnName: m.collectionChoices.copyTasks[i].source, targetCollectionsColumnName: m.collectionChoices.copyTasks[i].target}
		tableData = append(tableData, rowData)
	}

	m.collectionChoices.copyTaskTable = m.collectionChoices.copyTaskTable.WithRows(buildRows(tableData))
}

// AddRow adds a new row to the table
func (t *TableData) AddRow(row map[string]interface{}) {
	t.Rows = append(t.Rows, row)
}
