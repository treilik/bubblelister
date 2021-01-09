package boxer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
)

// Leave are the Boxers which holds the content Models
type Leave struct {
	Content     tea.Model
	Border      bool
	Width       int
	Heigth      int
	innerHeigth int
	innerWidth  int
	id          int
	Focus       bool

	N, NW, W, SW, S, SO, O, NO string
}

func NewLeave() Leave {
	return Leave{
		Border: true,
		N:      "─",
		NW:     "╭",
		W:      "│",
		SW:     "╰",
		NO:     "╮",
		O:      "│",
		SO:     "╯",
		S:      "─",
	}
}

func (l Leave) Init() tea.Cmd {
	return l.Content.Init()
}
func (l Leave) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// account for Border width/heigth
	switch msg := msg.(type) {
	case LeaveMsg:
		if msg.LeaveID == l.id {
			l.Focus = msg.Focus
		}
	case tea.WindowSizeMsg:
		l.Width = msg.Width
		l.Heigth = msg.Height
		if !l.Border {
			l.innerWidth = msg.Width
			l.innerHeigth = msg.Height
			return l.Content.Update(msg)
		}
		l.innerHeigth = msg.Height - strings.Count(l.N+l.S, "\n") - 2
		l.innerWidth = msg.Width - ansi.PrintableRuneWidth(l.W+l.O)
		newContent, cmd := l.Content.Update(tea.WindowSizeMsg{Height: l.innerHeigth, Width: l.innerWidth})
		l.Content = newContent
		return l, cmd
	}
	if !l.Focus {
		return l, nil
	}
	newContent, cmd := l.Content.Update(msg)
	l.Content = newContent
	return l, cmd
}
func (l Leave) View() string {
	lines, err := l.lines()
	if err != nil {
		return err.Error()
	}
	return strings.Join(lines, "\n")
}

// Lines returns the fully rendert (maybe with borders) lines and fullfills the Boxer Interface
func (l Leave) Lines() ([]string, error) {
	return l.lines()
}

func (l *Leave) lines() ([]string, error) {
	boxer, ok := l.Content.(Boxer)
	var lines []string
	var err error
	if ok {
		lines, err = boxer.Lines()
	} else {
		lines = strings.Split(l.Content.View(), "\n")
	}
	if length := len(lines); length > l.innerHeigth {
		return nil, NewProporationError(l)
	}

	// expand to match heigth
	if len(lines) < l.innerHeigth {
		lines = append(lines, make([]string, l.innerHeigth-len(lines))...)
	}
	for i, line := range lines {
		lineWidth := ansi.PrintableRuneWidth(line)
		if lineWidth > l.innerWidth {
			return nil, NewProporationError(l)
		}
		lines[i] = line + strings.Repeat(" ", l.innerWidth-lineWidth)
	}

	if !l.Border {
		return lines, err
	}
	fullLines := make([]string, 0, len(lines)+strings.Count(l.N, "\n")+1)
	fullLines = append(fullLines, l.NW+strings.Repeat(l.N, l.innerWidth)+l.NO)
	for _, line := range lines {
		fullLines = append(fullLines, l.W+line+l.O)
	}
	fullLines = append(fullLines, l.SW+strings.Repeat(l.S, l.innerWidth)+l.SO)

	return fullLines, err
}
