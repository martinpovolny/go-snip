package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	edookit "github.com/martinpovolny/go-snip/edookit"
)

func main() {
	mock := flag.Bool("mock", false, "run with mock data (no API required)")
	flag.Parse()

	// Load .env from cwd (ignore if missing)
	_ = godotenv.Load()

	baseURL := os.Getenv("API_URL")
	user := os.Getenv("API_USER")
	pass := os.Getenv("API_PASSWORD")
	// API_AUTH_MODE: currently only "basic" is supported
	_ = os.Getenv("API_AUTH_MODE")

	var source edookit.DataSource
	useMock := *mock || baseURL == ""

	if useMock {
		source = &MockSource{}
		if baseURL == "" {
			baseURL = "demo.edookit.net"
		}
	} else {
		source = edookit.New(baseURL, user, pass)
	}

	m := newModel(source, baseURL, useMock)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
