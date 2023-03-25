package bubblelister

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

// Model is a bubbletea List of strings
type Model struct {
	listItems []item

	LessFunc   func(fmt.Stringer, fmt.Stringer) bool // function used for sorting
	EqualsFunc func(fmt.Stringer, fmt.Stringer) bool // used after sorting, to be set from the user

	// offset or margin between the cursor and the visible border
	CursorOffset int

	// The visible Area size of the list
	Width, Height int

	cursorIndex int

	// The maximal amout of lines (not items) infront of the cursor index
	lineOffset int

	// Wrap changes the number of lines which get displayed. 0 means unlimited lines.
	Wrap int

	PrefixGen Prefixer
	SuffixGen Suffixer

	LineStyle    termenv.Style
	CurrentStyle termenv.Style

	// mutex for unique ids
	idMutex   *sync.Mutex
	idCounter int
}

// NewModel returns a Model with some save/sane defaults
// design to transfer as much internal information to the user
func NewModel() Model {
	// just reverse colors to keep there information
	curStyle := termenv.Style{}.Reverse()
	var mut sync.Mutex
	return Model{
		// Try to keep $CursorOffset lines between Cursor and screen Border
		CursorOffset: 5,
		lineOffset:   5,

		// show all lines
		Wrap: 0,

		// show line number
		PrefixGen: NewPrefixer(),

		CurrentStyle: curStyle,

		idMutex: &mut,
	}
}

// Init does nothing
func (m Model) Init() tea.Cmd {
	return nil
}

// View renders the List output according to the current model
// and returns "empty" if the list has no items. This might change in the future.
func (m Model) View() string {

	lines, err := m.lines()
	if err != nil {
		return err.Error()
	}

	return strings.Join(lines, "\n")
}

// Update only handles WindowSizeMsg, everything else has to be implemented by the user.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// Lines renders the visible lines of the list
// by calling the String Methodes of the items
// and if present the pre- and suffix function.
// If there is not enough space, or there a no
// item within the list, nil and a error is returned.
func (m Model) Lines() ([]string, error) {
	return m.lines()
}

// lines is a method which gets called by View and Lines,
// because the are functions and if the View would call the Lines function directly,
// the model would be copied twice, once for the View call and ones for the Lines call.
// But since they both (Lines and View) can call this method,
// its only one copy of the model when calling either View or Lines.
func (m *Model) lines() ([]string, error) {
	if m.Len() == 0 {
		return nil, NoItems(fmt.Errorf("no items within the list"))
	}
	// check visible area
	if m.Height <= 0 || m.Width <= 0 {
		return nil, fmt.Errorf("Can't display with zero width or hight of Viewport")
	}

	linesBefor := make([]string, 0, m.lineOffset)
	// loop to add the item(-lines) befor the cursor to the return lines
	// dont add cursor item
	for c := 1; m.cursorIndex-c >= 0 && c <= m.Height; c++ {
		index := m.cursorIndex - c
		// Get the Width of each suf/prefix
		var prefixWidth, suffixWidth int
		if m.PrefixGen != nil {
			prefixWidth = m.PrefixGen.InitPrefixer(m.listItems[index].value, c, m.cursorIndex, m.lineOffset, m.Width, m.Height)
		}
		if m.SuffixGen != nil {
			suffixWidth = m.SuffixGen.InitSuffixer(m.listItems[index].value, c, m.cursorIndex, m.lineOffset, m.Width, m.Height)
		}
		// Get actual content width
		contentWidth := m.Width - prefixWidth - suffixWidth

		// Check if there is space for the content left
		if contentWidth <= 0 {
			return nil, fmt.Errorf("Can't display with zero width or hight of Viewport")
		}
		itemLines, _ := m.getItemLines(index, contentWidth)
		// append lines in revers order
		for i := len(itemLines) - 1; i >= 0 && len(linesBefor) < m.lineOffset; i-- {
			linesBefor = append(linesBefor, itemLines[i])
		}
	}

	// append lines (befor cursor) in correct order to allLines
	allLines := make([]string, 0, m.Height)
	for c := len(linesBefor) - 1; c >= 0; c-- {
		allLines = append(allLines, linesBefor[c])
	}

	// Handle list items, start at cursor and go till end of list or visible (break)
	for index := m.cursorIndex; index < m.Len() && index < m.cursorIndex+m.Height; index++ {
		// Get the Width of each suf/prefix
		var prefixWidth, suffixWidth int
		if m.PrefixGen != nil {
			prefixWidth = m.PrefixGen.InitPrefixer(m.listItems[index].value, index, m.cursorIndex, m.lineOffset, m.Width, m.Height)
		}
		if m.SuffixGen != nil {
			suffixWidth = m.SuffixGen.InitSuffixer(m.listItems[index].value, index, m.cursorIndex, m.lineOffset, m.Width, m.Height)
		}
		// Get actual content width
		contentWidth := m.Width - prefixWidth - suffixWidth

		// Check if there is space for the content left
		if contentWidth <= 0 {
			return nil, fmt.Errorf("Can't display with zero width or hight of Viewport")
		}
		itemLines, _ := m.getItemLines(index, contentWidth)
		// append lines in correct order
		for i := 0; i < len(itemLines) && len(allLines) < m.Height; i++ {
			allLines = append(allLines, itemLines[i])
		}
	}
	if len(allLines) == 0 {
		return nil, fmt.Errorf("no visible lines")
	}

	return allLines, nil
}

// NoItems is a error returned when the list is empty
type NoItems error

// NotFound gets return if the search does not yield a result
type NotFound error

// OutOfBounds is return if an index is outside the list boundary's
type OutOfBounds error

// MultipleMatches gets return if the search yield more result
type MultipleMatches error

// ConfigError is return if there is a error with the configuration of the list Model
type ConfigError error

// NilValue is returned if there was a request to set nil as value of a list item.
type NilValue error

// UnhandledKey is returned when there is no binding for this key press.
type UnhandledKey error

//TODO make New functions for errors and Messages

// ValidIndex returns a error when the list has no items or the index is out of bounds.
// And the nearest valid index in case of OutOfBounds error, else the index it self and no error.
func (m *Model) ValidIndex(index int) (int, error) {
	if m.Len() <= 0 {
		return 0, NoItems(fmt.Errorf("the list has no items"))
	}
	if index < 0 {
		return 0, OutOfBounds(fmt.Errorf("the requested index (%d) is in front the list begin (%d)", index, 0))
	}
	if index > m.Len()-1 {
		return m.Len() - 1, OutOfBounds(fmt.Errorf("the requested index (%d) is beyond the list end (%d)", index, m.Len()-1))
	}
	return index, nil
}

func (m *Model) validOffset(newCursor int) (int, error) {
	if m.CursorOffset*2 > m.Height {
		return 0, ConfigError(fmt.Errorf("CursorOffset must be less than have the screen height"))
	}
	newCursor, err := m.ValidIndex(newCursor)
	if m.Len() <= 0 {
		return m.CursorOffset, err
	}
	amount := newCursor - m.cursorIndex
	if amount == 0 {
		if m.lineOffset < m.CursorOffset {
			return m.CursorOffset, nil
		}
		return m.lineOffset, nil
	}
	newOffset := m.lineOffset + amount

	if m.Wrap != 1 {
		// assume down (positive) movement
		start := 0
		stop := amount - 1 // exclude target item (-lines)

		d := 1
		if amount < 0 {
			d = -1
			stop = amount * d
			start = 1 // exclude old cursor position
		}

		var lineSum int
		for i := start; i <= stop; i++ {
			lineSum += len(m.itemLines(m.listItems[m.cursorIndex+i*d], m.cursorIndex+i*d))
		}
		newOffset = m.lineOffset + lineSum*d
	}

	if newOffset < m.CursorOffset {
		newOffset = m.CursorOffset
	} else if newOffset > m.Height-m.CursorOffset-1 {
		newOffset = m.Height - m.CursorOffset - 1
	}
	return newOffset, err
}

// MoveCursor moves the cursor by amount and returns the absolut index of the cursor after the movement.
// If any error occurs the cursor is not moved.
func (m *Model) MoveCursor(amount int) (int, error) {
	target := m.cursorIndex + amount

	target, err := m.ValidIndex(target)
	if err != nil || amount == 0 {
		return target, err
	}
	newOffset, err := m.validOffset(target)
	if err != nil {
		return target, err
	}

	m.cursorIndex = target
	m.lineOffset = newOffset
	return target, nil
}

// SetCursor set the cursor to the specified index if possible,
// but If any error occurs the cursor is not moved.
func (m *Model) SetCursor(target int) (int, error) {
	target, err := m.ValidIndex(target)
	newOffset, _ := m.validOffset(target)
	if err != nil {
		return target, err
	}
	if target == m.cursorIndex {
		return target, nil
	}

	m.cursorIndex = target
	m.lineOffset = newOffset
	return target, nil
}

// Top moves the cursor to the first item if the list is not empty,
// else the cursor is not moved.
func (m *Model) Top() error {
	_, err := m.ValidIndex(0)
	if err != nil {
		return err
	}
	if m.cursorIndex == 0 {
		return nil
	}
	m.cursorIndex = 0
	m.lineOffset = m.CursorOffset
	return nil
}

// Bottom moves the cursor to the last item if the list is not empty,
// else the cursor is not moved.
func (m *Model) Bottom() error {
	end := len(m.listItems) - 1
	_, err := m.ValidIndex(end)
	if err != nil {
		return err
	}
	if m.cursorIndex == end {
		return nil
	}
	m.lineOffset = m.Height - m.CursorOffset
	m.SetCursor(end)
	return nil
}

// AddItems adds the given Items to the end of the list. Run Sort() afterwards, if you want to keep the list sorted.
// If entrys of itemList are nil they will not be added, and a NilValue error is returned.
// If you add very many Items, the program will get slower, since bubbletea is a elm architektur,
// Update and View are functions and are call with a copy of the list-Model which takes more time if the Model/List is bigger.
func (m *Model) AddItems(itemList ...fmt.Stringer) error {
	if len(itemList) == 0 {
		return nil
	}
	oldLenght := m.Len()
	for _, i := range itemList {
		if i == nil {
			continue
		}

		m.listItems = append(m.listItems, item{
			value: i,
			id:    m.getID(),
		})
	}
	if m.Len() < oldLenght+len(itemList) {
		err := NilValue(fmt.Errorf("there where '%d' nil values which where not added", m.Len()-oldLenght+len(itemList)))
		return err
	}
	return nil
}

// ResetItems replaces all list items with the new items, if a entry is nil its not added.
// If equals function is set and a new item yields true in comparison to the old cursor item
// the cursor is set on this (or if equals-func is bad the last-)item.
func (m *Model) ResetItems(newStringers ...fmt.Stringer) error {
	oldCursorItem, _ := m.GetCursorItem()
	// Reset Cursor
	m.cursorIndex = 0

	newItems := make([]item, 0, len(newStringers))
	for i, newValue := range newStringers {
		if newValue == nil {
			continue
		}
		newItems = append(newItems, item{value: newValue, id: m.getID()})

		if m.EqualsFunc != nil && oldCursorItem != nil && m.EqualsFunc(oldCursorItem, newValue) {
			m.cursorIndex = i
		}
	}

	m.listItems = newItems

	// reset LineOffset if Cursor was not set by matching through equals
	if m.cursorIndex == 0 {
		m.lineOffset = m.CursorOffset
	}
	// only sort if user set less function
	if m.LessFunc != nil {
		m.Sort()
	}
	return nil
}

// RemoveIndex removes and returns the item at the given index if it exists, else a error is returned.
// If the index matches the current cursor position the numeric of the cursor does not change except the list becomes to short, than the cursor will be at the end.
// The cursor will stay on the item it was on and thus the numeric position may change if the removed item was before the cursor item.
func (m *Model) RemoveIndex(index int) (fmt.Stringer, error) {
	if _, err := m.ValidIndex(index); err != nil {
		return nil, err
	}

	// exclude requested index/item
	var rest []item
	itemValue, _ := m.GetItem(index)
	if index+1 < m.Len() {
		rest = m.listItems[index+1:]
	}
	m.listItems = append(m.listItems[:index], rest...)

	// stay on the same item
	if index < m.cursorIndex {
		m.cursorIndex--
	}

	// check if cursor position is still valid and change if not
	oldCursor := m.cursorIndex
	newCursor, err := m.ValidIndex(oldCursor)
	newOffset, _ := m.validOffset(newCursor)
	m.cursorIndex = newCursor
	m.lineOffset = newOffset

	return itemValue, err
}

// Sort sorts the list items according to the set less-function or, if not set, after String comparison.
// Internally the sort.Sort interface is used, so this is not guaranteed to be a stable sort.
// If you need stable sorting, sort the items your self and reset the list with them.
// While sorting the cursor item can not change, but the cursor index can.
func (m *Model) Sort() {
	if m.Len() < 1 {
		return
	}
	old := m.listItems[m.cursorIndex].id
	sort.Sort(m)
	for i, item := range m.listItems {
		if item.id == old {
			m.cursorIndex = i
			break
		}
	}
	return
}

func (m *Model) Less(i, j int) bool {
	// If User does not provide less function use string comparison, but dont change m.less, to be able to see when user set one.
	if m.LessFunc == nil {
		return m.listItems[i].value.String() < m.listItems[j].value.String()
	}
	return m.LessFunc((m.listItems)[i].value, m.listItems[j].value)
}

func (m *Model) Swap(i, j int) {
	m.listItems[i], (m.listItems)[j] = m.listItems[j], (m.listItems)[i]
}

// Len returns the amount of list-items.
func (m *Model) Len() int {
	return len(m.listItems)
}

// MoveCursorItemTo moves the current cursor item to the index 'to'.
// If the target position does not exist a error is returned.
// The Cursor stays on the same item.
func (m *Model) MoveCursorItemTo(to int) error {
	i, err := m.GetCursorIndex()
	if err != nil {
		return err
	}
	return m.MoveCursorItemBy(to - i)
}

// MoveCursorItemBy moves the current cursor item RELATIV.
// If the target position does not exist a error is returned.
// The Cursor stays on the same item.
func (m *Model) MoveCursorItemBy(amount int) error {
	i, err := m.GetCursorIndex()
	if err != nil {
		return err
	}
	return m.MoveItemBy(i, amount)
}

// MoveItemTo moves the item at 'from' to the index 'to'.
// If the target position does not exist a error is returned.
// The Cursor stays on the same item.
func (m *Model) MoveItemTo(from, to int) error {
	i, err := m.GetCursorIndex()
	if err != nil {
		return err
	}
	return m.MoveItemBy(from, to-i)
}

// MoveItemBy moves the item at index RELATIV by amount to the end of the list.
// If the target position does not exist a error is returned.
// The Cursor stays on the same item.
func (m *Model) MoveItemBy(index, amount int) error {
	// check valid source
	if _, err := m.ValidIndex(index); err != nil {
		return err
	}
	// check valid target
	target, err := m.ValidIndex(index + amount)
	if err != nil {
		return err
	}

	// nothing to do
	if amount == 0 {
		return nil
	}

	if amount < 0 { // amount negative

		// since m.listItems is a slice make a deep copy for middle and rest so that they are independent from m.listItems
		middleSlice := m.listItems[target:index]
		middle := make([]item, len(middleSlice))
		copy(middle, middleSlice)

		var rest []item
		if index+1 < m.Len() {
			restSlice := m.listItems[index+1:]
			rest = make([]item, len(restSlice))
			copy(rest, restSlice)
		}
		// add beginning and moving item
		m.listItems = append(m.listItems[:target], m.listItems[index])
		// add the middle and the rest
		m.listItems = append(m.listItems, middle...)
		m.listItems = append(m.listItems, rest...)

	} else { // amount positive

		// since m.listItems is a slice make a deep copy for middle and rest so that they are independent from m.listItems
		middleSlice := m.listItems[index+1 : target+1]
		middle := make([]item, len(middleSlice))
		copy(middle, middleSlice)

		var rest []item
		if target+1 < m.Len() {
			restSlice := m.listItems[target+1:]
			rest = make([]item, len(restSlice))
			copy(rest, restSlice)
		}

		movingItem := m.listItems[index]

		// add middle
		m.listItems = append(m.listItems[:index], middle...)
		// add the moving item and the rest
		m.listItems = append(m.listItems, movingItem)
		m.listItems = append(m.listItems, rest...)
	}

	// keep cursor visible
	linOff, _ := m.validOffset(target)
	m.lineOffset = linOff

	if m.cursorIndex == index {
		// change cursor position to stay on the same item
		m.cursorIndex = target
		return nil
	}

	// change cursor position on the same item
	if target < m.cursorIndex && index >= m.cursorIndex {
		m.cursorIndex++
	}
	if target > m.cursorIndex && index <= m.cursorIndex {
		m.cursorIndex--
	}
	return nil
}

// GetIndex returns NotFound error if the Equals Method is not set (SetEquals) or no item is found
// else it returns the index of the found item
func (m *Model) GetIndex(toSearch fmt.Stringer) (int, error) {
	if m.EqualsFunc == nil {
		return -1, NotFound(fmt.Errorf("no equals function provided. Use SetEquals to set it"))
	}
	tmpList := m.listItems
	matchList := make([]chan bool, len(tmpList))
	equ := m.EqualsFunc

	for i, item := range tmpList {
		resChan := make(chan bool)
		matchList[i] = resChan
		go func(f, s fmt.Stringer, equ func(fmt.Stringer, fmt.Stringer) bool, res chan<- bool) {
			res <- equ(f, s)
		}(item.value, toSearch, equ, resChan)
	}

	var c, lastIndex int
	for i, resChan := range matchList {
		if <-resChan {
			c++
			lastIndex = i
		}
	}
	if c > 1 {
		// TODO performance: trust User and remove check for multiple matches?
		return -c, MultipleMatches(fmt.Errorf("The provided equals function yields multiple matches betwen one and other fmt.Stringer's"))
	}
	if c == 0 {
		return -1, NotFound(fmt.Errorf("No item found"))
	}
	return lastIndex, nil
}

// UpdateItem takes a index and updates the item at the index with the given function
// or if index outside the list returns OutOfBounds error.
// If the returned fmt.Stringer value is nil, then the item gets removed from the list.
// If you want to keep the list sorted run Sort() after updating a item.
// if the update function returns a error, the item is not changed and the error is directly returned
func (m *Model) UpdateItem(index int, updater func(fmt.Stringer) (fmt.Stringer, error)) error {
	index, err := m.ValidIndex(index)
	if err != nil {
		return err
	}
	v, err := updater(m.listItems[index].value)
	if err != nil {
		return err
	}

	// remove item when value equals nil
	if v == nil {
		_, err = m.RemoveIndex(index)
		return err
	}
	m.listItems[index].value = v
	return nil
}

// GetCursorIndex returns the current cursor position within the List,
// or a NoItems error if the list has no items on which the cursor could be.
func (m *Model) GetCursorIndex() (int, error) {
	if m.Len() == 0 {
		return 0, NoItems(fmt.Errorf("the list has no items on which the cursor could be"))
	}
	return m.cursorIndex, nil
}

// GetCursorItem returns the item at the current cursor position within the List
// or a NoItems error if the list has no items on which the cursor could be.
func (m *Model) GetCursorItem() (fmt.Stringer, error) {
	if m.Len() == 0 {
		return nil, NoItems(fmt.Errorf("the list has no items on which the cursor could be"))
	}
	return m.listItems[m.cursorIndex].value, nil
}

// GetItem returns the item if the index exists otherwise a error.
func (m *Model) GetItem(index int) (fmt.Stringer, error) {
	index, err := m.ValidIndex(index)
	if err != nil {
		return nil, err
	}
	return m.listItems[index].value, nil
}

// GetAllItems returns all items in the list in current order
func (m *Model) GetAllItems() []fmt.Stringer {
	list := m.listItems
	stringerList := make([]fmt.Stringer, len(list))
	for i, item := range list {
		stringerList[i] = item.value
	}
	return stringerList
}

// getID returns a new for this list unique id
// to identify the items and set the cursor after sorting correctly.
func (m *Model) getID() int {
	m.idMutex.Lock()
	// skip the 0 to be able to distinguish valid and default ids
	m.idCounter++
	m.idMutex.Unlock()
	return m.idCounter
}
