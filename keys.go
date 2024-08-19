package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up               key.Binding
	Down             key.Binding
	Left             key.Binding
	Right            key.Binding
	Enter            key.Binding
	Help             key.Binding
	Quit             key.Binding
	Filter           key.Binding
	QuitFilter       key.Binding
	Select           key.Binding
	IncreasePageSize key.Binding
	DecreasePageSize key.Binding
	IncreaseRows     key.Binding
	DecreaseRows     key.Binding
	DeleteCopyTask   key.Binding
	ToggleAltView    key.Binding
	StartCopy        key.Binding
	EditCopyTasks    key.Binding
}

type keyModel struct {
	keys       keyMap
	inputStyle lipgloss.Style
	quitting   bool
}

// add key bindings
var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "page left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "page right"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/  ", "filter (start)"),
	),
	QuitFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "filter (exit)"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	IncreasePageSize: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "Page size (+1)"),
	),
	DecreasePageSize: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "Page size (-1)"),
	),
	IncreaseRows: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "Rows (+10)"),
	),
	DecreaseRows: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "Row (-10)"),
	),
	DeleteCopyTask: key.NewBinding(
		key.WithKeys("del"),
		key.WithHelp("del", "Delete Copy Choice"),
	),
	StartCopy: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Start Copy"),
	),
	ToggleAltView: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "Toggle view"),
	),
}

func (m model) databaseChoicesHelp() string {
	pad := lipgloss.NewStyle().Padding(2, 2)
	seperator := ": "
	navigation := subtleStyle.Render(m.keyBindings.keys.Up.Help().Key+seperator+m.keyBindings.keys.Up.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Down.Help().Key+seperator+m.keyBindings.keys.Down.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Left.Help().Key+seperator+m.keyBindings.keys.Left.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Right.Help().Key+seperator+m.keyBindings.keys.Right.Help().Desc) + "\n"

	table := subtleStyle.Render(m.keyBindings.keys.IncreasePageSize.Help().Key+seperator+m.keyBindings.keys.IncreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreasePageSize.Help().Key+seperator+m.keyBindings.keys.DecreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.IncreaseRows.Help().Key+seperator+m.keyBindings.keys.IncreaseRows.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreaseRows.Help().Key+seperator+m.keyBindings.keys.DecreaseRows.Help().Desc) + "\n"

	other := subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Select.Help().Key+seperator+m.keyBindings.keys.Select.Help().Desc) + "\n"

	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(navigation)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(table)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(other)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, help...)
}

func (m model) collectionChoicesHelp() string {
	pad := lipgloss.NewStyle().Padding(2, 2)
	seperator := ": "
	navigation := subtleStyle.Render(m.keyBindings.keys.Up.Help().Key+seperator+m.keyBindings.keys.Up.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Down.Help().Key+seperator+m.keyBindings.keys.Down.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Left.Help().Key+seperator+m.keyBindings.keys.Left.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Right.Help().Key+seperator+m.keyBindings.keys.Right.Help().Desc) + "\n"

	table := subtleStyle.Render(m.keyBindings.keys.IncreasePageSize.Help().Key+seperator+m.keyBindings.keys.IncreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreasePageSize.Help().Key+seperator+m.keyBindings.keys.DecreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.IncreaseRows.Help().Key+seperator+m.keyBindings.keys.IncreaseRows.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreaseRows.Help().Key+seperator+m.keyBindings.keys.DecreaseRows.Help().Desc) + "\n"

	other := subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n"

	copy := subtleStyle.Render(m.keyBindings.keys.ToggleAltView.Help().Key+seperator+m.keyBindings.keys.ToggleAltView.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Select.Help().Key+seperator+m.keyBindings.keys.Select.Help().Desc) + "\n"

	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(navigation)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(table)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(other)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(copy)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, help...)
}

func (m model) collectionChoicesCopyHelp() string {
	pad := lipgloss.NewStyle().Padding(2, 2)
	seperator := ": "
	navigation := subtleStyle.Render(m.keyBindings.keys.Up.Help().Key+seperator+m.keyBindings.keys.Up.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Down.Help().Key+seperator+m.keyBindings.keys.Down.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Left.Help().Key+seperator+m.keyBindings.keys.Left.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Right.Help().Key+seperator+m.keyBindings.keys.Right.Help().Desc) + "\n"

	table := subtleStyle.Render(m.keyBindings.keys.IncreasePageSize.Help().Key+seperator+m.keyBindings.keys.IncreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreasePageSize.Help().Key+seperator+m.keyBindings.keys.DecreasePageSize.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.IncreaseRows.Help().Key+seperator+m.keyBindings.keys.IncreaseRows.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DecreaseRows.Help().Key+seperator+m.keyBindings.keys.DecreaseRows.Help().Desc) + "\n"

	other := subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n"

	copy := subtleStyle.Render(m.keyBindings.keys.ToggleAltView.Help().Key+seperator+m.keyBindings.keys.ToggleAltView.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.StartCopy.Help().Key+seperator+m.keyBindings.keys.StartCopy.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.DeleteCopyTask.Help().Key+seperator+m.keyBindings.keys.DeleteCopyTask.Help().Desc) + "\n"

	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(navigation)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(table)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(other)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(copy)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, help...)
}
