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
	ToggleAltView    key.Binding
	StartCopy        key.Binding
	EditCopyTasks    key.Binding
	Restart          key.Binding
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
		key.WithHelp("i", "page size (+1)"),
	),
	DecreasePageSize: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "page size (-1)"),
	),
	StartCopy: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "start Copy"),
	),
	ToggleAltView: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle view"),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
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
		subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n"

	other := subtleStyle.Render(m.keyBindings.keys.Select.Help().Key+seperator+m.keyBindings.keys.Select.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n"

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
		subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n"

	copy := subtleStyle.Render(m.keyBindings.keys.ToggleAltView.Help().Key+seperator+"view selections") + "\n" +
		"\n" +
		subtleStyle.Render(m.keyBindings.keys.Select.Help().Key+seperator+m.keyBindings.keys.Select.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n"

	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(navigation)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(table)),
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
		subtleStyle.Render(m.keyBindings.keys.Filter.Help().Key+seperator+m.keyBindings.keys.Filter.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.QuitFilter.Help().Key+seperator+m.keyBindings.keys.QuitFilter.Help().Desc) + "\n"

	copy := subtleStyle.Render(m.keyBindings.keys.ToggleAltView.Help().Key+seperator+"view collections") + "\n" +
		subtleStyle.Render(m.keyBindings.keys.StartCopy.Help().Key+seperator+m.keyBindings.keys.StartCopy.Help().Desc) + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Select.Help().Key+seperator+"remove") + "\n" +
		subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n"

	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(navigation)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(table)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(copy)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, help...)
}

func (m model) RestartHelp() string {
	pad := lipgloss.NewStyle().Padding(2, 2)
	seperator := ": "

	quit := subtleStyle.Render(m.keyBindings.keys.Quit.Help().Key+seperator+m.keyBindings.keys.Quit.Help().Desc) + "\n"
	restart := subtleStyle.Render(m.keyBindings.keys.Restart.Help().Key+seperator+m.keyBindings.keys.Restart.Help().Desc) + "\n"
	help := []string{
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(restart)),
		lipgloss.JoinVertical(lipgloss.Center, pad.Render(quit)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, help...)
}
