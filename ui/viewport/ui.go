package viewport

import (
	"fmt"
	"github.com/blackmann/gurl/lib"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"net/http"
)

type keymap struct {
	nextTab     key.Binding
	previousTab key.Binding
}

type Model struct {
	activeTab    int
	keybinds     keymap
	tabs         []string
	responseBody string
	enabled      bool

	height int

	// state
	headers http.Header

	// tabs
	responseModel
	requestBodyModel
	headersModel
	responseHeadersModel
}

func NewViewport() Model {
	tabs := []string{"Headers (:q)", "Request Body (:w)", "Response (:e)", "Response Headers (:r)"}
	keybinds := keymap{
		nextTab:     key.NewBinding(key.WithKeys("shift+f"), key.WithHelp("⌥→", "Next tab")),
		previousTab: key.NewBinding(key.WithKeys("shift+b"), key.WithHelp("⌥←", "Prev tab")),
	}

	return Model{
		tabs:                 tabs,
		keybinds:             keybinds,
		responseHeadersModel: newResponseHeadersModel(),
		headers:              http.Header{},
	}
}

func (model *Model) SetResponse(response lib.Response) tea.Msg {
	return response
}

func (model *Model) SetEnabled(enabled bool) {
	model.enabled = enabled
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, model.keybinds.nextTab):
			model.activeTab = (model.activeTab + 1) % len(model.tabs)
			return model, nil

		case key.Matches(msg, model.keybinds.previousTab):
			var newTab int
			if model.activeTab == 0 {
				newTab = 2
			} else {
				newTab = model.activeTab - 1
			}
			model.activeTab = newTab
		}

	case lib.FreeText:
		switch msg {
		case ":q":
			model.activeTab = 0

		case ":w":
			model.activeTab = 1

		case ":e":
			model.activeTab = 2

		case ":r":
			model.activeTab = 3
		}

		return model, nil

	case tea.WindowSizeMsg:
		renderHeight := msg.Height - 3
		resizeMsg := tea.WindowSizeMsg{Height: renderHeight, Width: msg.Width}

		model.responseModel, _ = model.responseModel.Update(resizeMsg)
		model.requestBodyModel, _ = model.requestBodyModel.Update(resizeMsg)
		model.headersModel, _ = model.headersModel.Update(resizeMsg)
		model.responseHeadersModel, _ = model.responseHeadersModel.Update(resizeMsg)

		return model, nil

	case headerItem:
		model.headers.Set(msg.key, msg.value)

		cmd := func() tea.Msg {
			return requestHeaders(model.headers)
		}

		return model, cmd
	}

	var cmds []tea.Cmd

	model.responseModel, cmd = model.responseModel.Update(msg)
	cmds = append(cmds, cmd)

	model.requestBodyModel, cmd = model.requestBodyModel.Update(msg)
	cmds = append(cmds, cmd)

	model.headersModel, cmd = model.headersModel.Update(msg)
	cmds = append(cmds, cmd)

	model.responseHeadersModel, cmd = model.responseHeadersModel.Update(msg)
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	viewportStyle := lipgloss.NewStyle().Height(model.height)

	styledTabs := make([]string, len(model.tabs))

	for i, tab := range model.tabs {
		if i != model.activeTab {
			styledTabs = append(styledTabs, inactiveTabStyle.Render(tab))
		} else {
			styledTabs = append(styledTabs, activeTabStyle.Render(tab))
		}
	}

	tabsRow := tabGroupStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, styledTabs...))

	content := ""

	switch model.activeTab {
	case 0:
		content = model.headersModel.View()
	case 1:
		content = model.requestBodyModel.View()
	case 2:
		content = model.responseModel.View()
	case 3:
		content = model.responseHeadersModel.View()
	}

	return viewportStyle.Render(fmt.Sprintf("%s\n%s", tabsRow, content))
}

func (model *Model) GetHeaders() http.Header {
	return nil
}

func (model *Model) GetBody() io.Reader {
	return nil
}
