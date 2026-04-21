package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	checkedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	statusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("237"))
)
