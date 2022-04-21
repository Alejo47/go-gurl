package statusbar

import (
	"fmt"
	"github.com/blackmann/gurl/lib"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// A message sent with a value to set as the status' value
type statusUpdate lib.Status

type Model struct {
	spinner  spinner.Model
	spinning bool

	width        int
	status       lib.Status
	commandEntry string
	message      lib.ShortMessage
	mode         lib.Mode
}

func NewStatusBar() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{spinner: s}
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commandInput:
		model.commandEntry = string(msg)
		return model, nil

	case statusUpdate:
		// Allow to flow through so ticking can begin for
		// status == PROCESSING
		model.status = lib.Status(msg)

	case tea.WindowSizeMsg:
		model.width = msg.Width
		return model, nil

	case lib.ShortMessage:
		model.message = msg
		return model, nil

	case lib.Mode:
		model.mode = msg
		return model, nil
	}

	if model.status == lib.PROCESSING {
		if !model.spinning {
			model.spinning = true
			return model, model.spinner.Tick
		}

		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)
		return model, cmd
	} else {
		model.spinning = false
	}

	return model, nil
}

func (model Model) View() string {
	var status string

	switch model.status {
	case lib.PROCESSING:
		status = fmt.Sprintf("%s Processing", model.spinner.View())
	case lib.IDLE:
		status = neutralStatusStyle.Render("Idle")

	case lib.ERROR:
		status = errorStatusStyle.Render("Error")

	default:
		value := model.status.GetValue()

		if value < 300 {
			status = okStatusStyle.Render(fmt.Sprintf("%d", value))
		} else if value < 400 {
			status = okStatusStyle.Render(fmt.Sprintf("%d", value))
		} else {
			status = errorStatusStyle.Render(fmt.Sprintf("%d", value))
		}

	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999"))

	half := lipgloss.NewStyle().Width(model.width/2 - 2) // Left/right padding removed

	rightHalf := half.Copy().Align(lipgloss.Right).
		Render(fmt.Sprintf("%s :%s", model.commandEntry, mutedStyle.Render(string(model.mode))))

	leftHalf := half.Copy().Align(lipgloss.Left).
		Render(fmt.Sprintf("%s %s", status, mutedStyle.Render(string(model.message))))

	render := fmt.Sprintf("%s %s", leftHalf, rightHalf)

	return barStyle.Copy().Width(model.width).Render(render)
}
