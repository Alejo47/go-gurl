package main

import (
	"fmt"
	"github.com/blackmann/gurl/common/appcmd"
	"github.com/blackmann/gurl/common/request"
	"github.com/blackmann/gurl/common/status"
	"github.com/blackmann/gurl/ui/addressbar"
	"github.com/blackmann/gurl/ui/statusbar"
	"github.com/blackmann/gurl/ui/viewport"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log"
	"net/http"
	"os"
)

type keymap struct {
	nextTab           key.Binding
	quit              key.Binding
	toggleCommandMode key.Binding
}

func getDefaultKeyBinds() keymap {
	return keymap{
		nextTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("^tab", "Toggle Regions")),

		quit: key.NewBinding(key.WithKeys("ctrl+c")),

		toggleCommandMode: key.NewBinding(key.WithKeys("esc")),
	}
}

type model struct {
	// Config
	keybinds keymap

	// Views
	addressBar addressbar.Model
	viewport   viewport.Model
	statusBar  statusbar.Model

	// State
	activeRegion int
	commandMode  bool
	command      string
}

func (m model) Init() tea.Cmd {
	return m.addressBar.Init()
}

func (m *model) resizeViewport(netHeight int, netWidth int) {
	statusBarHeight := lipgloss.Height(m.statusBar.View())
	addressBarHeight := lipgloss.Height(m.addressBar.View())

	m.viewport.SetHeight(netHeight - (statusBarHeight + addressBarHeight))
	m.viewport.Width = netWidth
	m.statusBar, _ = m.statusBar.Update(statusbar.Width(netWidth))
}

func (m model) submitRequest(address request.Address) tea.Cmd {
	return func() tea.Msg {
		res, err := http.Get(address.Url)

		if err != nil {
			log.Panicln("Error occurred", err)
			return nil
		}
		body, err := io.ReadAll(res.Body)

		if err != nil {
			log.Panicln("Error occured while reading response body", err)
		}

		return appcmd.Response{
			Body: string(body),
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.commandMode {
			if key.Matches(msg, m.keybinds.toggleCommandMode) {
				m.commandMode = false
				m.command = ""
				m.statusBar, _ = m.statusBar.Update(statusbar.CommandMsg(""))

				return m, nil
			}

			switch msg.Type {
			case tea.KeyRunes:
				m.command += string(msg.Runes[0])

			case tea.KeyBackspace:
				if len(m.command) > 0 {
					m.command = m.command[:len(m.command)-1]
				}

			case tea.KeyEnter:
				cmd := getFreeTextCommand(m.command)
				m.command = ""
				m.commandMode = false

				return m, cmd
			}

			var prefix string
			if m.commandMode {
				prefix = "> "
			} else {
				prefix = ""
			}

			m.statusBar, _ = m.statusBar.Update(statusbar.CommandMsg(fmt.Sprintf("%s%s", prefix, m.command)))

			return m, nil
		}

		switch {
		case key.Matches(msg, m.keybinds.nextTab):
			m.activeRegion = (m.activeRegion + 1) % 2 // only two views
			return m, nil

		case key.Matches(msg, m.keybinds.toggleCommandMode):
			m.commandMode = true
			m.statusBar, _ = m.statusBar.Update(statusbar.CommandMsg(">"))
			return m, nil

		case key.Matches(msg, m.keybinds.quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		// TODO: Rewrite
		m.resizeViewport(msg.Height, msg.Width)
		return m, nil

	case appcmd.FreeText:
		m.viewport, _ = m.viewport.Update(msg)
		return m, nil

	case appcmd.Trigger:
		switch msg {
		case appcmd.NewRequest:
			var cmds []tea.Cmd
			var cmd tea.Cmd

			m.statusBar, cmd = m.statusBar.Update(statusbar.StatusMsg(status.PROCESSING))
			cmds = append(cmds, cmd)

			// TODO: Use .Update
			m.viewport.SetEnabled(false)

			cmds = append(cmds, m.submitRequest(m.addressBar.GetAddress()))

			return m, tea.Batch(cmds...)
		}

	case appcmd.Response:
		// TODO: Use .Update
		m.viewport.SetResponse(msg.Body)
		m.viewport.SetEnabled(true)

		m.statusBar, _ = m.statusBar.Update(statusbar.StatusMsg(status.IDLE))
		return m, nil
	}

	var cmds []tea.Cmd

	// Forward the unhandled command to the active region
	switch m.activeRegion {
	case 0:
		var cmd tea.Cmd
		m.addressBar, cmd = m.addressBar.Update(msg)

		cmds = append(cmds, cmd)

	case 1:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)

		cmds = append(cmds, cmd)
	}

	// For ticks
	var cmd tea.Cmd
	m.statusBar, cmd = m.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s\n%s",
		m.addressBar.View(),
		m.viewport.View(),
		m.statusBar.View())
}

func newAppModel() model {
	return model{
		addressBar: addressbar.NewAddressBar(),
		keybinds:   getDefaultKeyBinds(),
		statusBar:  statusbar.NewStatusBar(),
		viewport:   viewport.NewViewport(),
	}
}

func getFreeTextCommand(cmd string) tea.Cmd {
	return func() tea.Msg {
		return appcmd.FreeText(cmd)
	}
}

func main() {
	// Set up logger
	f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Panicln("Failed to open log file")
	}

	defer f.Close()

	log.SetOutput(f)

	// Initialize and start app
	app := tea.NewProgram(newAppModel(), tea.WithAltScreen())

	if err := app.Start(); err != nil {
		log.Panicln("Error occurred", err)
	}
}
