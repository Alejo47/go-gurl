package addressbar

import (
	"errors"
	"github.com/blackmann/gurl/lib"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type Model struct {
	input textinput.Model
}

func NewAddressBar() Model {
	t := textinput.New()
	t.Placeholder = "/GET @adeton/shops"
	t.Prompt = "¬ "

	t.Focus()

	return Model{
		input: t,
	}
}

func (model Model) Init() tea.Cmd {
	textinput.Blink()
	return textinput.Blink
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return model, lib.SubmitNewRequest
		}
	}

	var cmd tea.Cmd
	model.input, cmd = model.input.Update(msg)

	return model, cmd
}

func (model Model) View() string {
	return model.input.View()
}

func (model Model) GetAddress() (lib.Address, error) {
	trimmedAddr := strings.Trim(model.input.Value(), " ")

	if len(trimmedAddr) == 0 {
		return lib.Address{}, errors.New("")
	}

	// Expecting at least one part and most 2
	// When two, the first part is the method and the latter is the endpoint
	parts := strings.Split(trimmedAddr, " ")

	method := "GET"
	endpoint := parts[len(parts)-1]

	if len(parts) > 1 {
		method = parts[0]
	}

	return lib.Address{Method: strings.ToUpper(method), Url: endpoint}, nil
}
