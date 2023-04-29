package bubblelister

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/ansi"
	"github.com/treilik/reflow/wordwrap"
)

// Item are Items used in the list Model
// to hold the Content represented as a string
type item struct {
	value fmt.Stringer
	id    int
}

// itemLines returns the lines of the item string value wrapped to the according content-width
// and the write amount of lines accoring to m.Wrap
func (m *Model) itemLines(i item, index int) []string {
	var preWidth, sufWidth int
	if m.PrefixGen != nil {
		preWidth = m.PrefixGen.InitPrefixer(i.value, index, m.cursorIndex, m.lineOffset, m.Width, m.Height)
	}
	if m.SuffixGen != nil {
		sufWidth = m.SuffixGen.InitSuffixer(i.value, index, m.cursorIndex, m.lineOffset, m.Width, m.Height)
	}
	contentWith := m.Width - preWidth - sufWidth
	// TODO hard limit the string length
	lines := strings.Split(wordwrap.HardWrap(i.value.String(), contentWith, "    "), "\n")
	if m.Wrap != 0 && len(lines) > m.Wrap {
		return lines[:m.Wrap]
	}
	return lines
}

// getItemLines surrounds the line content with the according prefix and suffix
func (m *Model) getItemLines(index, contentWidth int) ([]string, error) {
	_, err := m.ValidIndex(index)
	if err != nil {
		return nil, err
	}
	item := m.listItems[index]
	lines := m.itemLines(item, index)
	lenLines := len(lines)
	completLines := make([]string, lenLines)

	for c := 0; c < lenLines; c++ {
		lineContent := lines[c]
		// Surrounding content
		var linePrefix, lineSuffix string
		if m.PrefixGen != nil {
			linePrefix = m.PrefixGen.Prefix(c, lenLines)
		}
		if m.SuffixGen != nil {
			free := contentWidth - ansi.PrintableRuneWidth(lineContent)
			suffix := m.SuffixGen.Suffix(c, lenLines)
			if suffix != "" {
				lineSuffix = fmt.Sprintf("%s%s", strings.Repeat(" ", free), suffix)
			}
		}

		// Join all
		line := fmt.Sprintf("%s%s%s", linePrefix, lineContent, lineSuffix)

		// Highlighting of current item lines
		style := m.LineStyle
		if index == m.cursorIndex {
			style = m.CurrentStyle
		}

		// Highlight and write line
		completLines[c] = style.Styled(line)
	}
	return completLines, nil
}

// StringItem is just a convenience to satisfy the fmt.Stringer interface with plain strings
type StringItem string

func (s StringItem) String() string {
	return string(s)
}

// MakeStringerList is a shortcut to convert a string List to a List that satisfies the fmt.Stringer Interface
func MakeStringerList(list ...string) []fmt.Stringer {
	stringerList := make([]fmt.Stringer, len(list))
	for i, item := range list {
		stringerList[i] = StringItem(item)
	}
	return stringerList
}
