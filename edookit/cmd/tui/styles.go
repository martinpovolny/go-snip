package main

import "github.com/charmbracelet/lipgloss"

// Edookit purple palette
const (
	colorPurple    = "#7C3AED"
	colorPurpleDim = "#5B21B6"
	colorLavender  = "#C4B5FD"
	colorBg        = "#1E1B4B"
	colorBgRow     = "#0F0D23"
	colorSelected  = "#4C1D95"
	colorText      = "#E2E8F0"
	colorDim       = "#94A3B8"
	colorGreen     = "#34D399"
	colorRed       = "#F87171"
	colorYellow    = "#FCD34D"
)

var (
	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorPurple)).
			Foreground(lipgloss.Color(colorText)).
			Bold(true).
			Padding(0, 1)

	headerDimStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorPurple)).
			Foreground(lipgloss.Color(colorLavender)).
			Padding(0, 1)

	titleBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorPurpleDim)).
			Foreground(lipgloss.Color(colorLavender)).
			Bold(true).
			Padding(0, 1)

	titleCountStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorPurpleDim)).
			Foreground(lipgloss.Color(colorDim)).
			Padding(0, 1)

	hintsStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgRow)).
			Foreground(lipgloss.Color(colorDim)).
			Padding(0, 1)

	hintKeyStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgRow)).
			Foreground(lipgloss.Color(colorLavender)).
			Bold(true)

	statusOkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGreen))

	statusErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorRed))

	cmdPromptStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBg)).
			Foreground(lipgloss.Color(colorYellow)).
			Bold(true).
			Padding(0, 1)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPurpleDim))
)

func hint(key, desc string) string {
	return hintKeyStyle.Render(key) + hintsStyle.Render(" "+desc)
}
