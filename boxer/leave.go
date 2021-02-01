package boxer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

// Leave are the Boxers which holds the content Models
type Leave struct {
	Content     tea.Model
	Border      bool
	BorderStyle termenv.Style
	Width       int
	Heigth      int
	innerHeigth int
	innerWidth  int
	Focus       bool
	id          int

	N, NW, W, SW, S, SO, O, NO string
}

// NewLeave returns a leave with border enabled and set
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

// Init is a proxy to the Content Init
func (l Leave) Init() tea.Cmd {
	return l.Content.Init()
}

// Update takes care about the seting of the id of this leave
// and the changing of the WindowSizeMsg depending on the border
// and the focus style of the border.
func (l Leave) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO Remove hardcoded Focus styling:
	if l.Focus {
		l.BorderStyle = termenv.String().Foreground(termenv.ColorProfile().Color("#00aaaa"))
	}
	if !l.Focus {
		l.BorderStyle = termenv.String()
	}

	switch msg := msg.(type) {
	case InitIDs:
		// make a uniq channel for this leave
		idGen := make(chan int)
		// send this uniq channel over the general channel
		msg <- idGen
		// to recive a uniq id from the uniq channel
		l.id = <-idGen
		return l, nil
	case FocusLeave:
		if !l.Focus {
			return l, nil
		}
		// TODO is there always a node befor a leave?
		for c := len(msg.path) - 1; c >= 0; c-- {
			if msg.path[c].vertical == msg.vertical {
				l.Focus = false                  // TODO make sure a leave is allways focused
				l.BorderStyle = termenv.String() // TODO remove hardcoding of style
				newIndex := msg.path[c].index - 1
				if msg.next {
					newIndex = msg.path[c].index + 1
				}
				if newIndex < 0 {
					// TODO
				}
				if len(msg.path) > 2 {
					//panic(msg.path) //TODO
				}

				// update the msg to be the path for the new focus1
				msg.path = msg.path[:c+1]    // exclude the rest of the path since its not valid for the new path
				msg.path[c].index = newIndex // TODO her is no check possible if out of bound
				// return the new path (from the root till the changing index) in side a ChangeFocus to signal the new node that one of its children should take the focus.

				return l, func() tea.Msg { return ChangeFocus{focus: true, newFocus: msg} }
			}
		}
	case ChangeFocus:
		l.Focus = msg.focus
		return l, func() tea.Msg { return nil } // TODO why is a cmd nessecary to redraw in time?
	case tea.WindowSizeMsg:
		l.Width = msg.Width
		l.Heigth = msg.Height
		if !l.Border {
			l.innerWidth = msg.Width
			l.innerHeigth = msg.Height
			return l.Content.Update(msg)
		}
		// account for Border width/heigth
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

// View is used to satisfy the tea.Model interface and returnes either the joined lines
// or the Error string if err of lines is not nil.
func (l Leave) View() string {
	lines, err := l.lines()
	if err != nil {
		return err.Error()
	}
	return strings.Join(lines, "\n")
}

// Lines returns the fully rendert (maybe with borders) lines and fullfills the Boxer Interface.
// Error may be of type ProporationError.
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

	// expand to match width
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
	// draw border
	fullLines := make([]string, 0, len(lines)+strings.Count(l.N, "\n")+1)
	firstLine := l.NW + strings.Repeat(l.N, l.innerWidth) + l.NO
	begin, end := l.W, l.O
	lastLine := l.SW + strings.Repeat(l.S, l.innerWidth) + l.SO

	// if set style border
	if l.BorderStyle.String() != "" {
		firstLine = l.BorderStyle.Styled(firstLine)
		begin = l.BorderStyle.Styled(begin)
		end = l.BorderStyle.Styled(end)
		lastLine = l.BorderStyle.Styled(lastLine)
	}

	fullLines = append(fullLines, firstLine)
	for _, line := range lines {
		fullLines = append(fullLines, begin+line+end)
	}
	fullLines = append(fullLines, lastLine)

	return fullLines, err
}
