package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	progressBarWidth = 71
	dotChar          = " • "
	banner           = `
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

type (
	getDatabasesMsg   []string
	getCollectionsMsg []string
)

type errMsg struct {
	err     error
	context string
}

type collectionChoice struct {
	Id       int
	Name     string
	Spinner  spinner.Model
	Complete bool
}

type FatalError struct {
	Text    string
	Context string
}

type model struct {
	Databases         []string           // Databases on server
	DatabaseChoice    int                // Database chosen by user
	DatabaseChosen    bool               // Has user made database selection
	Collections       []string           // Collections in database
	CurrentCollection int                // Collection cursor is current on
	CollectionChoices []collectionChoice // Collections user has selected
	CollectionsChosen bool               // Has user made database selection
	Quitting          bool               // Has user quit application
	Storage           storage            // Storage
	FatalError        *FatalError        // Fatal Error details
}

// For messages that contain errors
func (e errMsg) Error() string { return e.err.Error() }

// Commands

func (m model) getDatabases() tea.Msg {

	databases, err := m.Storage.getDatabases()

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return errMsg{err, "getting databases"}
	}

	return getDatabasesMsg(databases)

}

func (m model) getCollections() tea.Msg {
	collections, err := m.Storage.getCollections(m.Databases[m.DatabaseChoice])

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return errMsg{err, "getting collections"}
	}

	return getCollectionsMsg(collections)
}

func (m model) doCopy() []tea.Cmd {
	var err error
	var cmds []tea.Cmd

	for _, c := range m.CollectionChoices {
		cmd := func() tea.Msg {
			err = m.Storage.copy(m.Collections[c.Id], m.Databases[m.DatabaseChoice])
			if err != nil {
				return errMsg{err, "copying records"}
			}

			return copyMsg{collectionId: c.Id}
		}

		cmds = append(cmds, cmd)
	}

	return cmds
}

func (m model) Init() tea.Cmd {
	return m.getDatabases
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
	case getDatabasesMsg:
		m.Databases = msg
		return m, tea.ClearScreen
	case getCollectionsMsg:
		m.Collections = msg
		return m, tea.ClearScreen
	case copyMsg:
		for i := 0; i < len(m.CollectionChoices); i++ {
			if msg.collectionId == m.CollectionChoices[i].Id {
				m.CollectionChoices[i].Complete = true
			}
		}
		return m, nil
	case spinner.TickMsg:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)
		for i := 0; i < len(m.CollectionChoices); i++ {
			m.CollectionChoices[i].Spinner, cmd = m.CollectionChoices[i].Spinner.Update(msg)
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
	if !m.DatabaseChosen {
		return updateDatabaseChoices(msg, m)
	} else if !m.CollectionsChosen {
		return updateCollectionChoices(msg, m)
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
	} else if !m.DatabaseChosen {
		s = databaseChoicesView(m)
	} else if !m.CollectionsChosen {
		s = collectionChoicesView(m)
	} else {
		s = copyView(m)
	}
	return mainStyle.Render("\n" + s + "\n\n")
}

// Update functions

// Update loop for the first view where you're choosing a database.
func updateDatabaseChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.DatabaseChoice++
			if m.DatabaseChoice >= len(m.Databases) {
				m.DatabaseChoice = len(m.Databases) - 1
			}
		case "up":
			m.DatabaseChoice--
			if m.DatabaseChoice < 0 {
				m.DatabaseChoice = 0
			}
		case "enter":
			m.DatabaseChosen = true
			return m, m.getCollections
		}
	}

	return m, nil
}

// Update loop for the first view where you're choosing collections.
func updateCollectionChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.CurrentCollection++
			if m.CurrentCollection >= len(m.Collections) {
				m.CurrentCollection = len(m.Collections) - 1
			}
		case "up":
			m.CurrentCollection--
			if m.CurrentCollection < 0 {
				m.CurrentCollection = 0
			}
		case " ":
			if !containsChoice(m.CollectionChoices, m.CurrentCollection) {
				var s = spinner.New()
				s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
				s.Spinner = spinner.Line

				collectionChoice := collectionChoice{
					Id:      m.CurrentCollection,
					Name:    m.Collections[m.CurrentCollection],
					Spinner: s}
				m.CollectionChoices = append(m.CollectionChoices, collectionChoice)
			} else {
				for i := 0; i < len(m.CollectionChoices); i++ {
					if m.CollectionChoices[i].Id == m.CurrentCollection {
						m.CollectionChoices = RemoveChoice(m.CollectionChoices, i)
					}
				}
			}
			return m, nil
		case "enter":

			if len(m.CollectionChoices) == 0 {
				return m, nil
			}

			m.CollectionsChosen = true

			var cmds []tea.Cmd
			var c = m.doCopy()
			cmds = append(cmds, c...)

			for i := 0; i < len(m.CollectionChoices); i++ {
				cmd := func() tea.Msg {
					return m.CollectionChoices[i].Spinner.Tick()
				}

				cmds = append(cmds, cmd)
			}

			return m, tea.Batch(cmds...)
		}
	}

	return m, nil
}

// Views

// The error view
func errorView(m model) string {
	tpl := "\nA fatal error occured while %s\n"
	tpl += "%s\n%s\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	return fmt.Sprintf(tpl, m.FatalError.Context, errorImage, "Error Message: "+keywordStyle.Render(m.FatalError.Text))
}

// The first view, where you're choosing a database
func databaseChoicesView(m model) string {
	tpl := banner + "\n"
	tpl += "Choose the source database\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.Databases {
		choices += fmt.Sprintf("%s\n", checkbox(choice, m.DatabaseChoice == i))
	}

	return fmt.Sprintf(tpl, choices)
}

// The second view, where you're choosing a collections
func collectionChoicesView(m model) string {
	tpl := banner + "\n"
	tpl += "Choose one or more collections to copy all records to target\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("up/down: select") + dotStyle +
		subtleStyle.Render("space: choose") + dotStyle +
		subtleStyle.Render("enter: start copy") + dotStyle +
		subtleStyle.Render("q, esc: quit")

	var choices string
	for i, choice := range m.Collections {
		checked := multiCheckbox(choice, containsChoice(m.CollectionChoices, i), m.CurrentCollection == i)
		choices += fmt.Sprintf("%s\n", checked)
	}

	return fmt.Sprintf(tpl, choices)
}

// The copy view shown after a collections has been chosen
func copyView(m model) string {
	var progress, label string
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	tpl := banner + "\n"
	tpl += fmt.Sprintf("Copying data to target database (%s)\n", keywordStyle.Render(m.Databases[m.DatabaseChoice]))
	tpl += "%s\n\n\n"
	tpl += subtleStyle.Render("q, esc: quit")

	for i := 0; i < len(m.CollectionChoices); i++ {

		label = "Copying..."
		spinner := fmt.Sprintf("%s %s", m.CollectionChoices[i].Spinner.View(), textStyle(label))
		if m.CollectionChoices[i].Complete {
			label = "Done"
			spinner = fmt.Sprintf("%s %s", "✅", textStyle(label))
		}
		progress += "\n\n" + keywordStyle.Render(m.CollectionChoices[i].Name) + " - " + spinner
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

// Checkbox used when user can select multiple
func multiCheckbox(label string, checked bool, current bool) string {
	if checked {
		return checkboxStyle.Render("[x] " + label)
	}
	// Is cursor over checkbox
	if current {
		return checkboxStyle.Render("[ ] " + label)
	} else {
		return fmt.Sprintf("[ ] %s", label)
	}

}

// Utils

// Check if collection exists in selected choices slice.
func containsChoice(slice []collectionChoice, id int) bool {
	for _, v := range slice {
		if v.Id == id {
			return true
		}
	}
	return false
}

// Remove collection from collection choices slice.
func RemoveChoice(s []collectionChoice, id int) []collectionChoice {
	ret := make([]collectionChoice, 0)
	ret = append(ret, s[:id]...)
	return append(ret, s[id+1:]...)
}
