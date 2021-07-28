package list

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

// TestLines test if the models Lines methode returns the write amount of lines
func TestEmptyLines(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should do nothing") // yet
	}
	m.Height = 50
	m.Width = 80
	_, err := m.Lines()
	if err == nil {
		t.Error("A list with no entrys should return a error.")
	}
	m.Sort()
	_, err = m.Lines()
	if err == nil {
		t.Error("A list with no entrys should return a error.")
	}
}

// TestBasicsLines test lines without linebreaks and with content shorter than the max content-width.
func TestBasicsLines(t *testing.T) {
	m := NewModel()
	m.Height = 50
	m.Width = 80
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()

	m.Wrap = 1

	// Check Cursor position
	if i, err := m.GetCursorIndex(); i != 0 || err == nil {
		t.Errorf("the cursor index of a new Model should be '0' and not: '%d' and there should be a error: %#v", i, err)
	}

	// first two swaped
	orgList := MakeStringerList("2", "1", "3", "4", "5", "6", "7", "8", "9")
	m.AddItems(orgList)

	m.MoveCursor(1)
	// Sort them
	m.Sort()
	// swap them again
	m.MoveItem(1)
	// should be the like the beginning
	sortedItemList := m.GetAllItems()

	if len(orgList) != len(sortedItemList) {
		t.Errorf("the list should not change size")
	}

	// Process/check all orgList
	for c, item := range orgList {
		if item.String() != sortedItemList[c].String() {
			t.Errorf("the old strings should match the new, but dont: %q, %q", item.String(), sortedItemList[c].String())
		}
	}

	m.Top()
	out, _ := m.Lines()
	if len(out) > 50 {
		t.Errorf("Lines should never have more (%d) lines than Screen has lines: %d", len(out), m.Height)
	}

	light := "\x1b[7m"
	cur := ">"
	sep := "╭"
	for i, line := range out {
		// Check Prefixes
		num := fmt.Sprintf("%d", i+1)
		prefix := light + strings.Repeat(" ", 2-len(num)) + num + sep + cur
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n%s\n with linenumber %d should be:\n%s\n", line, i, prefix)
		}
		cur = " "
		sep = "├"
		light = ""
	}
}

// TestWrappedLines test a simple case of many items with linebreaks.
func TestWrappedLines(t *testing.T) {
	m := NewModel()
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	m.Height = 50
	m.Width = 80
	m.AddItems(MakeStringerList("\n0", "1\n2", "3\n4", "5\n6", "7\n8"))

	out, _ := m.Lines()
	wrap, sep := "│", "├"
	num := "\x1b[7m  "
	for i := 1; i < len(out); i++ {
		line := out[i]
		if i%2 == 0 {
			num = fmt.Sprintf(" %1d", (i/2)+1)
		}
		if i%2 == 1 {
			sep = wrap
		}
		prefix := fmt.Sprintf("%s%s %d", num, sep, i-1)
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n'%s'\n with linenumber %d should be:\n'%s'\n", line, i, prefix)
		}
		num = "  "
		sep = "├"
	}
}

// TestMultiLineBreaks test one selected item
func TestMultiLineBreaks(t *testing.T) {
	m := NewModel()
	m.PrefixGen = NewPrefixer()
	m.SuffixGen = NewSuffixer()
	m.Height = 50
	m.Width = 80
	m.AddItems(MakeStringerList("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"))
	out, _ := m.Lines()
	prefix := "\x1b[7m 1╭>"
	for i, line := range out {
		if !strings.HasPrefix(line, prefix) {
			t.Errorf("The prefix of the line:\n'%s'\n with linenumber %d should be:\n'%s'\n", line, i, prefix)
		}
		prefix = "\x1b[7m  │ "
	}
}

// Movements
func TestMovementKeys(t *testing.T) {
	m := NewModel()
	m.Wrap = 1
	m.Height = 50
	m.Width = 80
	m.AddItems(MakeStringerList("\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"))

	start, finish := 0, 1
	_, err := m.MoveCursor(1)
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'MoveCursor(1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.cursorIndex)
	}
	start, finish = 15, 14
	m.cursorIndex = start
	_, err = m.MoveCursor(-1)
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'MoveCursor(-1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.cursorIndex)
	}

	start, finish = 55, 56
	m.cursorIndex = start
	err = m.MoveItem(1)
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'MoveItem(1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.cursorIndex)
	}
	m.lineOffset = 15
	start, finish = 15, 14
	m.cursorIndex = start
	err = m.MoveItem(-1)
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'MoveItem(-1)' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.cursorIndex)
	}
	if m.lineOffset != 14 {
		t.Errorf("up movement should change the Item offset to '14' but got: %d", m.lineOffset)
	}
	finish = m.Len() - 1
	err = m.Bottom()
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'Bottom()' should have nil error but got: '%#v' and move the Cursor to last index: '%d', but got: %d", err, m.Len()-1, m.cursorIndex)
	}
	finish = 0
	m.cursorIndex = start
	err = m.Top()
	if m.cursorIndex != finish || err != nil {
		t.Errorf("'Top()' should have nil error but got: '%#v' and move the Cursor to index '%d', but got: %d", err, finish, m.cursorIndex)
	}
	_, err = m.SetCursor(10)
	if m.cursorIndex != 10 || err != nil {
		t.Errorf("SetCursor should set the cursor to index '10' but gut '%d' and err should be nil but got '%s'", m.cursorIndex, err)
	}
}

// WindowMsg
func TestWindowMsg(t *testing.T) {
	m := NewModel()
	width, height := 80, 50

	newModel, cmd := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m, _ = newModel.(Model)

	if cmd != nil {
		t.Errorf("comand should be nil and not: '%#v'", cmd)
	}
	if m.Width != width {
		t.Errorf("the Width should be %#v and not: %#v", width, m.Width)
	}
	if m.Height != height {
		t.Errorf("the Width should be %#v and not: %#v", height, m.Height)
	}

}

// TestGetIndex sets a equals function and searches After the index of a specific item with GetIndex
func TestGetIndex(t *testing.T) {
	m := NewModel()
	_, err := m.GetIndex(StringItem("z"))
	if err == nil {
		t.Errorf("Get Index should return a error but got nil")
	}
	m.AddItems(MakeStringerList("a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"))
	m.EqualsFunc = func(a, b fmt.Stringer) bool { return a.String() == b.String() }
	index, err := m.GetIndex(StringItem("z"))
	if err != nil {
		t.Errorf("GetIndex should not return a err: %s", err)
	}
	if index != m.Len()-1 {
		t.Errorf("GetIndex returns wrong index: '%d' instead of '%d'", index, m.Len()-1)
	}
}

// TestWithinBorder test if indexes are within the listborders
func TestWithinBorder(t *testing.T) {
	m := NewModel()
	_, err := m.ValidIndex(0)
	if _, ok := err.(NoItems); !ok {
		t.Errorf("a empty list has no item '0', should return a NoItems error, but got: %#v", err)
	}
}

// TestCopy test if if Copy returns a deep copy
func TestCopy(t *testing.T) {
	org := NewModel()
	sec := org.Copy()

	org.LessFunc = func(a, b fmt.Stringer) bool { return a.String() < b.String() }

	if &org == sec {
		t.Errorf("Copy should return a deep copy but has the same pointer:\norginal: '%p', copy: '%p'", &org, sec)
	}

	if fmt.Sprintf("%#v", org.listItems) != fmt.Sprintf("%#v", sec.listItems) ||

		// All should be the same except the changed less function
		fmt.Sprintf("%p", org.LessFunc) == fmt.Sprintf("%p", sec.LessFunc) ||
		fmt.Sprintf("%p", org.EqualsFunc) != fmt.Sprintf("%p", sec.EqualsFunc) ||

		fmt.Sprintf("%#v", org.CursorOffset) != fmt.Sprintf("%#v", sec.CursorOffset) ||

		fmt.Sprintf("%#v", org.Width) != fmt.Sprintf("%#v", sec.Width) ||
		fmt.Sprintf("%#v", org.Height) != fmt.Sprintf("%#v", sec.Height) ||
		fmt.Sprintf("%#v", org.cursorIndex) != fmt.Sprintf("%#v", sec.cursorIndex) ||
		fmt.Sprintf("%#v", org.lineOffset) != fmt.Sprintf("%#v", sec.lineOffset) ||

		fmt.Sprintf("%#v", org.Wrap) != fmt.Sprintf("%#v", sec.Wrap) ||

		fmt.Sprintf("%#v", org.PrefixGen) != fmt.Sprintf("%#v", sec.PrefixGen) ||
		fmt.Sprintf("%#v", org.SuffixGen) != fmt.Sprintf("%#v", sec.SuffixGen) ||

		fmt.Sprintf("%#v", org.LineStyle) != fmt.Sprintf("%#v", sec.LineStyle) ||
		fmt.Sprintf("%#v", org.CurrentStyle) != fmt.Sprintf("%#v", sec.CurrentStyle) {

		t.Errorf("Copy should have same string repesentation except different less function pointer:\n orginal: '%#v'\n    copy: '%#v'", org, sec)
	}
}

// TestSetCursor tests if the LineOffset and Cursor positions are correct
func TestSetCursor(t *testing.T) {
	m := NewModel()
	m.Height = 50
	m.Width = 80
	m.AddItems(MakeStringerList("\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n", ""))
	type test struct {
		oldLineOffset  int
		oldCursorIndex int
		target         int
		newLineOffset  int
		newCursorIndex int
	}
	toTest := []test{
		// forwards
		{0, 0, -2, 0, 0}, // wrong request -> no change
		{0, 0, 2, 5, 2},
		{0, 4, 8, 8, 8},
		{0, 5, 0, 5, 0},
		{0, 0, 19, 38, 19},
		{0, 0, 25, 44, 25},
		{0, 0, 100, 0, 0}, // wrong request -> no change
		// backwards
		{45, m.Len() - 1, -2, 45, m.Len() - 1}, // wrong request -> no change
		{45, m.Len() - 1, 2, 5, 2},
		{45, m.Len() - 1, 8, 5, 8},
		{45, m.Len() - 1, 0, 5, 0},
		{45, m.Len() - 1, 19, 5, 19},
		{45, m.Len() - 1, 25, 5, 25},
		{45, m.Len() - 1, 100, 45, m.Len() - 1}, // wrong request -> no change
	}
	for i, tCase := range toTest {
		m.cursorIndex = tCase.oldCursorIndex
		m.lineOffset = tCase.oldLineOffset
		m.SetCursor(tCase.target)
		if m.cursorIndex != tCase.newCursorIndex {
			t.Errorf("In Test number: %d, the returned cursor index is wrong:\n'%#v' and should be:\n'%#v' after requesting target: %d", i, m.cursorIndex, tCase.newCursorIndex, tCase.target)
		}
		if m.lineOffset != tCase.newLineOffset {
			t.Errorf("In Test number: %d, the returned Line Offset is wrong:\n'%#v' and should be:\n'%#v' after requesting target: %d", i, m.lineOffset, tCase.newLineOffset, tCase.target)
		}
	}
}

// TestMoveItem test wrong arguments
func TestMoveItem(t *testing.T) {
	m := NewModel()
	err := m.MoveItem(0)
	err, ok := err.(OutOfBounds)
	if !ok {
		t.Errorf("MoveItem called on a empty list should return a OutOfBounds error, but got: %s", err)
	}
	m.AddItems(MakeStringerList(""))
	err = m.MoveItem(0)
	if ok && err != nil {
		t.Errorf("MoveItem(0) should not return a error on a not empty list, but got '%s'", err)
	}
	err = m.MoveItem(1)
	err, ok = err.(OutOfBounds)
	if !ok {
		t.Errorf("MoveItem should return a OutOfBounds error if traget is beyond list border, but got: '%s'", err)
	}
}

// TestView tests if View returns a String (of a returned lines)
func TestView(t *testing.T) {
	m := NewModel()
	if m.View() == "" {
		t.Error("View should never return a empty string since this does not update the screen") // TODO changed this in bubbletea
	}
	if _, err := m.Lines(); err != nil && m.View() != err.Error() {
		t.Error("if Lines returnes a error View should return the error string")
	}
	testStr := "test"
	m.AddItems(MakeStringerList(testStr, testStr))
	m.SetCursor(1)
	if _, err := m.Lines(); err == nil {
		t.Error("a none empty list should return a error when the screen is to small to displax anything")
	}
	m.Height = 10
	m.Width = 100
	if _, err := m.Lines(); err != nil || !strings.Contains(m.View(), testStr) {
		t.Errorf("a none empty list should not return a error but got:\n%sand the content should be within the returned string from View:\n%s", err, m.View())
	}
}

// TestRemoveIndex test if the item at the index was removed
func TestRemoveIndex(t *testing.T) {
	m := NewModel()
	item, err := m.RemoveIndex(0)
	if item != nil && err != nil {
		t.Error("RemoveIndex should return a error and a nil value when the index is not valid")
	}
	testStr := "test"
	m.AddItems(MakeStringerList(testStr))
	item, err = m.RemoveIndex(0)
	if item.String() != testStr && err != nil && m.Len() != 0 {
		t.Error("RemoveIndex should return no error and the corresponding string value when the index is valid")
	}
}

// TestResetItems test if list is replaced
func TestResetItems(t *testing.T) {
	m := NewModel()
	testStr := "test"
	m.AddItems(MakeStringerList(testStr))
	secondStr := "replaced"
	m.ResetItems(MakeStringerList(secondStr))
	if item, err := m.RemoveIndex(0); item.String() != secondStr || err == nil || m.Len() > 1 {
		t.Error("the list was not replaced, but should have been")
	}
}

// TestUpdateItem
func TestUpdateItem(t *testing.T) {
	m := NewModel()
	testStr := "test"
	m.AddItems(MakeStringerList(testStr))
	m.UpdateItem(0, func(fmt.Stringer) (fmt.Stringer, tea.Cmd) { return nil, nil })
	if item, err := m.RemoveIndex(0); item != nil || err == nil {
		t.Error("UpdateItem should return a command and the item should be deleted if the returned Stringer is nil")
	}
	m.AddItems(MakeStringerList(testStr))
	secondStr := "replaced"
	m.UpdateItem(0, func(fmt.Stringer) (fmt.Stringer, tea.Cmd) { return StringItem(secondStr), nil })
	if item, err := m.RemoveIndex(0); item.String() != secondStr || err == nil {
		t.Error("UpdateItem should return a command and the item should be replaced")
	}
}
