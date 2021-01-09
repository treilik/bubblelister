package boxer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
)

// Boxer is a interface to render multiple bubbles (within a tree) to the terminal screen.
type Boxer interface {
	Lines() ([]string, error)
	tea.Model
}

// Model is a bubble to manage/bundle other bubbles into boxes on the screen
type Model struct {
	Childs        []BoxerSize
	Height, Width int
	Stacked       bool
}

// BoxerSize holds a boxer value and the current size the box of this boxer should have
type BoxerSize struct {
	Boxer         Boxer
	Width, Heigth int
}

type ProportionError error

type LeaveMsg struct {
	LeaveID int
	Focus   bool
}

func NewProporationError(b Boxer) error {
	return fmt.Errorf("the Lines function of this boxer: '%v'\nhas returned to much or long lines", b)
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	cmdList := make([]tea.Cmd, len(m.Childs))
	for _, child := range m.Childs {
		cmdList = append(cmdList, child.Boxer.Init())
	}
	return tea.Batch(cmdList...)
}

// Update handles the ratios between the different Boxers
// though the according fanning of the WindowSizeMsg's
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmdList []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		default:
			for i, box := range m.Childs {
				newModel, cmd := box.Boxer.Update(msg)
				newBoxer, ok := newModel.(Boxer)
				if ok {
					box.Boxer = newBoxer
				}
				m.Childs[i] = box
				cmdList = append(cmdList, cmd)
			}
		}
	case tea.WindowSizeMsg:
		amount := len(m.Childs)
		for i, box := range m.Childs {
			newHeigth := msg.Height
			newWidth := (msg.Width) / amount
			if m.Stacked {
				newHeigth = (msg.Height) / amount
				newWidth = msg.Width
			}
			newModel, cmd := box.Boxer.Update(tea.WindowSizeMsg{Height: newHeigth, Width: newWidth})
			newBoxer, ok := newModel.(Boxer)
			if ok {
				box.Boxer = newBoxer
			} // TODO handle else case
			box.Heigth = newHeigth
			box.Width = newWidth
			m.Childs[i] = box
			cmdList = append(cmdList, cmd)
		}
	default:
		for i, box := range m.Childs {
			newModel, cmd := box.Boxer.Update(msg)
			newBoxer, ok := newModel.(Boxer)
			if ok {
				box.Boxer = newBoxer
			}
			m.Childs[i] = box
			cmdList = append(cmdList, cmd)
		}
	}
	return m, tea.Batch(cmdList...)
}

// View is only used for the top (root) node since all other Models use the Lines function.
func (m Model) View() string {
	lines, err := m.lines()
	if err != nil {
		return err.Error()
	}
	return strings.Join(lines, "\n")
}

// Lines returns the joined lines of all the contained Boxers
func (m Model) Lines() ([]string, error) {
	return m.lines()
}

// Lines returns the joined lines of all the contained Boxers
func (m *Model) lines() ([]string, error) {
	if m.Stacked {
		return verticalJoin(m.Childs)
	}
	return hotizontalJoin(m.Childs)
}

func hotizontalJoin(toJoin []BoxerSize) ([]string, error) {
	if len(toJoin) == 0 {
		return nil, fmt.Errorf("no childs to get lines from")
	}
	//            y  x
	var joinedStr [][]string
	var formerHeigth int
	for _, boxer := range toJoin {
		lines, err := boxer.Boxer.Lines()
		if err != nil {
			return nil, err
		}

		if len(lines) < boxer.Heigth {
			lines = append(lines, make([]string, boxer.Heigth-len(lines))...)
		}
		joinedStr = append(joinedStr, lines)
		if formerHeigth > 0 && formerHeigth != boxer.Heigth {
			return nil, fmt.Errorf("for horizontal join all have to be the same heigth") // TODO change to own error type
		}
		formerHeigth = boxer.Heigth
	}

	lenght := len(joinedStr)
	boxWidth := toJoin[0].Width
	var allStr []string
	// y
	for c := 0; c < formerHeigth; c++ {
		fullLine := make([]string, 0, lenght)
		// x
		for i := 0; i < lenght; i++ {
			line := joinedStr[i][c]
			lineWidth := ansi.PrintableRuneWidth(line)
			if lineWidth > boxWidth {
				return nil, NewProporationError(toJoin[i].Boxer)
			}
			var pad string
			if lineWidth < boxWidth {
				pad = strings.Repeat(" ", boxWidth-lineWidth)
			}
			fullLine = append(fullLine, line, pad)
		}
		allStr = append(allStr, strings.Join(fullLine, ""))
	}

	return allStr, nil
}

func verticalJoin(toJoin []BoxerSize) ([]string, error) {
	if len(toJoin) == 0 {
		return nil, fmt.Errorf("")
	}
	boxWidth := toJoin[0].Width
	var boxes []string
	var formerWidth int
	for _, child := range toJoin {
		if child.Boxer == nil {
			return nil, fmt.Errorf("cant work on nil Boxer") // TODO
		}
		lines, err := child.Boxer.Lines()
		if err != nil {
			return nil, err // TODO limit propagation of errors
		}
		if len(lines) > child.Heigth {
			return nil, NewProporationError(child.Boxer)
		}
		// check for  to wide lines and because we are on it, pad them to corrct width.
		for _, line := range lines {
			lineWidth := ansi.PrintableRuneWidth(line)
			if formerWidth > 0 && lineWidth != formerWidth {
				return nil, fmt.Errorf("for vertical join all boxes have to be the same width") // TODO change to own error type
			}
			line += strings.Repeat(" ", boxWidth-lineWidth)
		}
		boxes = append(boxes, lines...)
		// add more lines to boxes to match the Height of the child-box
		for c := 0; c < child.Heigth-len(lines); c++ {
			boxes = append(boxes, strings.Repeat(" ", boxWidth))
		}
	}
	return boxes, nil
}
