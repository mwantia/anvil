package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mwantia/anvil/internal/forge"
	"github.com/mwantia/anvil/internal/ui"
)

func main() {
	address := flag.String("address", "", "forge daemon address")
	token := flag.String("token", "", "bearer token")
	flag.Parse()

	client := forge.NewSDKClient(*address, *token)

	app := ui.NewApp(client)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "anvil:", err)
		os.Exit(1)
	}
}
