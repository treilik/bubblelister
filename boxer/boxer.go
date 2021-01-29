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
	tea.Model // TODO remove View
}

// Model is a bubble to manage/bundle other bubbles into boxes on the screen
type Model struct {
	children      []BoxSize
	Height, Width int
	Stacked       bool // TODO rename to vertical
	id            int

	errList []string // TODO remove?

	requestID chan<- chan int
}

// BoxSize holds a boxer value and the current size the box of this boxer should have
type BoxSize struct {
	Box           Boxer
	Width, Heigth int
}

// Start is a Msg to start the id spreading
type Start struct{}

// InitIDs is a Msg to spread the id's of the leaves
type InitIDs chan<- chan int

// ProportionError is for signaling that the string return by the View or Lines function has wrong proportions(width/height)
type ProportionError error

// FocusLeave is used to gather the path of each leave while its trasported to the leave.
type FocusLeave struct {
	path           []nodePos
	vertical, next bool
}

// ChangeFocus is the answere of FocusLeave and tells the parents to change the focus of the leaves by two msg.
type ChangeFocus struct {
	path  []nodePos
	focus bool
}

type nodePos struct {
	index    int
	vertical bool
	id       int //TODO remove
}

// NewProporationError returns a uniform string for this error
func NewProporationError(b Boxer) error {
	return fmt.Errorf("the Lines function of this boxer: '%v'\nhas returned to much or long lines", b)
}

// Init call the Init methodes of the Children and returns the batched/collected returned Cmd's of them
func (m Model) Init() tea.Cmd {
	cmdList := make([]tea.Cmd, len(m.children))
	for _, child := range m.children {
		cmdList = append(cmdList, child.Box.Init())
	}
	// the adding of the Start Msg leads to multiple Msg while only one is used and the rest gets ignored
	cmdList = append(cmdList, func() tea.Msg { return Start{} })
	return tea.Batch(cmdList...)
}

// Update handles the ratios between the different Boxers
// though the according fanning of the WindowSizeMsg's
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmdList []tea.Cmd
	switch msg := msg.(type) {
	case Start:
		// only the root node gets this all other ids will be set through the spreading of InitIDs
		// TODO should root node be a own struct? to handel the id spread-starting cleaner.
		if m.requestID != nil {
			return m, nil
		}
		m.id = m.getID()
		return m, func() tea.Msg { return InitIDs(m.requestID) }

	case InitIDs:
		if m.requestID == nil {
			m.requestID = msg
			genID := make(chan int)
			m.requestID <- genID
			m.id = <-genID
		}
		for i, box := range m.children {
			newModel, cmd := box.Box.Update(msg)
			newBoxer, ok := newModel.(Boxer)
			if !ok {
				continue
			}
			box.Box = newBoxer
			m.children[i] = box
			cmdList = append(cmdList, cmd)
		}
		return m, tea.Batch(cmdList...)

	// FocusLeave is a exception to the FAN-OUT of the Msg's because for each child there is a specific msg, similar to the WindowSizeMsg.
	case FocusLeave:
		for i, box := range m.children {
			// for each child append its position to the path
			newMsg := msg
			newMsg.path = append(msg.path, nodePos{index: i, vertical: m.Stacked, id: m.id})
			newModel, cmd := box.Box.Update(newMsg)
			// Focus
			newBoxer, ok := newModel.(Boxer)
			if !ok { // TODO
				continue
			}
			box.Box = newBoxer
			m.children[i] = box
			cmdList = append(cmdList, cmd)
		}
		return m, tea.Batch(cmdList...)

	// ChangedFocus is a exception to the FAN-OUT of the Msg's because its follows the specific path defined by the Msg-emitter.
	case ChangeFocus:
		var targetIndex int // TODO default is first of the array, so to speak most left and upper. should this be?
		if len(msg.path) > 0 {
			if ind := msg.path[0].index; ind >= len(m.children) || ind < 0 {
				return m, tea.Batch(func() tea.Msg { return fmt.Errorf("invalid path: %d", ind) }, func() tea.Msg { return ChangeFocus{focus: true} }) // TODO make error own type // by leaving the path in the ChangeFocus msg empty a default path/leave will be choosen.//TODO change order? first/(sequential) the focuschange to minemise the no focus time?
			}
			targetIndex = msg.path[0].index
		}
		childMsg := ChangeFocus{focus: msg.focus}
		if len(msg.path) > 1 {
			childMsg.path = msg.path[1:]
		}
		newModel, cmd := m.children[targetIndex].Box.Update(childMsg)
		var ok bool
		m.children[targetIndex].Box, ok = newModel.(Boxer)
		if !ok {
			panic("wrong type") // TODO
		}
		//return m, tea.Batch(cmd, func() tea.Msg { return fmt.Errorf("%v", msg.path) }) TODO
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "alt+right":
			// this is a exception to the FAN_OUT since the inital message does not becomes distributed but sparks the distributen of FocusLeave Msg's
			for i, box := range m.children {
				fMsg := FocusLeave{}
				fMsg.path = append(fMsg.path, nodePos{index: i, vertical: m.Stacked})
				fMsg.next = true
				fMsg.vertical = false
				newModel, cmd := box.Box.Update(fMsg)
				newBoxer, ok := newModel.(Boxer)
				if !ok {
					panic("wrong type") // TODO
				}
				box.Box = newBoxer
				m.children[i] = box
				cmdList = append(cmdList, cmd)
			}
			return m, tea.Batch(cmdList...)
		case "alt+left":
			// this is a exception to the FAN_OUT since the inital message does not becomes distributed but sparks the distributen of FocusLeave Msg's
			for i, box := range m.children {
				fMsg := FocusLeave{}
				fMsg.path = append(fMsg.path, nodePos{index: i, vertical: m.Stacked})
				fMsg.next = false
				fMsg.vertical = false
				newModel, cmd := box.Box.Update(fMsg)
				newBoxer, ok := newModel.(Boxer)
				if !ok {
					continue
				}
				box.Box = newBoxer
				m.children[i] = box
				cmdList = append(cmdList, cmd)
			}
			return m, tea.Batch(cmdList...)
		case "alt+up":
			// this is a exception to the FAN_OUT since the inital message does not becomes distributed but sparks the distributen of FocusLeave Msg's
			for i, box := range m.children {
				fMsg := FocusLeave{}
				fMsg.path = append(fMsg.path, nodePos{index: i, vertical: m.Stacked})
				fMsg.next = false
				fMsg.vertical = true
				newModel, cmd := box.Box.Update(fMsg)
				newBoxer, ok := newModel.(Boxer)
				if !ok {
					continue
				}
				box.Box = newBoxer
				m.children[i] = box
				cmdList = append(cmdList, cmd)
			}
			return m, tea.Batch(cmdList...)
		case "alt+down":
			// this is a exception to the FAN_OUT since the inital message does not becomes distributed but sparks the distributen of FocusLeave Msg's
			//for i, box := range m.children {
			fMsg := FocusLeave{}
			//fMsg.path = append(fMsg.path, nodePos{index: i, vertical: m.Stacked})
			fMsg.next = true
			fMsg.vertical = true
			//newModel, cmd := box.Box.Update(fMsg)
			//newBoxer, ok := newModel.(Boxer)
			//if !ok {
			//continue
			//}
			//box.Box = newBoxer
			//m.children[i] = box
			//cmdList = append(cmdList, cmd)
			//}
			//return m, tea.Batch(cmdList...)
			return m, func() tea.Msg { return fMsg }

		default:
			for i, box := range m.children {
				newModel, cmd := box.Box.Update(msg)
				newBoxer, ok := newModel.(Boxer)
				if !ok {
					continue
				}
				box.Box = newBoxer
				m.children[i] = box
				cmdList = append(cmdList, cmd)
			}
		}
		return m, tea.Batch(cmdList...)
	case tea.WindowSizeMsg:
		amount := len(m.children)
		for i, box := range m.children {
			newHeigth := msg.Height
			newWidth := (msg.Width) / amount
			if m.Stacked {
				newHeigth = (msg.Height) / amount
				newWidth = msg.Width
			}
			newModel, cmd := box.Box.Update(tea.WindowSizeMsg{Height: newHeigth, Width: newWidth})
			newBoxer, ok := newModel.(Boxer)
			if !ok {
				continue
			}
			box.Box = newBoxer
			box.Heigth = newHeigth
			box.Width = newWidth
			m.children[i] = box
			cmdList = append(cmdList, cmd)
		}
		return m, tea.Batch(cmdList...)
	case error:
		m.errList = append(m.errList, msg.Error())
		return m, nil
	default:
		for i, box := range m.children {
			newModel, cmd := box.Box.Update(msg)
			newBoxer, ok := newModel.(Boxer)
			if ok {
				box.Box = newBoxer
			}
			m.children[i] = box
			cmdList = append(cmdList, cmd)
		}
		return m, tea.Batch(cmdList...)
	}
}

// View is only used for the top (root) node since all other Models use the Lines function.
func (m Model) View() string {
	lines, err := m.lines()
	if err != nil {
		return err.Error()
	}
	return strings.Join(append(lines, m.errList...), "\n") // TODO make windows compatible
}

// Lines returns the joined lines of all the contained Boxers
func (m Model) Lines() ([]string, error) {
	return m.lines()
}

// Lines returns the joined lines of all the contained Boxers
func (m *Model) lines() ([]string, error) {
	if m.Stacked {
		return upDownJoin(m.children)
	}
	return leftRightJoin(m.children)
}

func leftRightJoin(toJoin []BoxSize) ([]string, error) {
	if len(toJoin) == 0 {
		return nil, fmt.Errorf("no children to get lines from")
	}
	//            y  x
	var joinedStr [][]string
	var formerHeigth int
	for _, boxer := range toJoin {
		lines, err := boxer.Box.Lines()
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
				return nil, NewProporationError(toJoin[i].Box)
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

func upDownJoin(toJoin []BoxSize) ([]string, error) {
	if len(toJoin) == 0 {
		return nil, fmt.Errorf("")
	}
	boxWidth := toJoin[0].Width
	var boxes []string
	var formerWidth int
	for _, child := range toJoin {
		if child.Box == nil {
			return nil, fmt.Errorf("cant work on nil Boxer") // TODO
		}
		lines, err := child.Box.Lines()
		if err != nil {
			return nil, err // TODO limit propagation of errors
		}
		if len(lines) > child.Heigth {
			return nil, NewProporationError(child.Box)
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

// AddChildren addes the given BoxerSize's as children
// but excludes nil-values and returns after adding the rest a Nil Error
func (m *Model) AddChildren(cList []BoxSize) error {
	var errCount int
	newChildren := make([]BoxSize, 0, len(cList))
	for _, newChild := range cList {
		switch c := newChild.Box.(type) {
		case Model:
			c.requestID = m.requestID
			newChild.Box = c
			newChildren = append(newChildren, newChild)
		case Leave:
			newChild.Box = c
			newChildren = append(newChildren, newChild)
		default:
			errCount++
		}
	}
	m.children = append(m.children, newChildren...)
	if errCount > 0 {
		return fmt.Errorf("%d entrys could not be added", errCount)
	}
	return nil
}

// getID returns a new for this Model(-tree) unique id
// to identify the nodes/leave and direct the message flow.
func (m *Model) getID() int {
	if m.requestID == nil {
		req := make(chan chan int)

		m.requestID = req

		// the id '0' is skiped to be able to distinguish zero-value and proper id TODO is this a valid/good way to go?
		go func(requ <-chan chan int) {
			for c := 2; true; c++ {
				send := <-requ
				send <- c
				close(send)
			}
		}(req)

		return 1
	}
	idChan := make(chan int)
	m.requestID <- idChan
	return <-idChan
}
