package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/boxer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	lf := list.NewModel()
	lf.AddItems(list.MakeStringerList([]string{"first"}))

	ls := list.NewModel()
	ls.AddItems(list.MakeStringerList([]string{"second"}))
	b := boxer.Model{}
	b.Childs = []boxer.BoxerSize{{
		Boxer: boxer.Boxer(lf),
	}, {
		Boxer: boxer.Boxer(ls),
	}}
	p := tea.NewProgram(b)
	if err := p.Start(); err != nil {
		fmt.Println("count not start program")
		os.Exit(1)
	}
}
