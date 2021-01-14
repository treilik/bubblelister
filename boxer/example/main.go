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
	lf.AddItems(list.MakeStringerList([]string{"first", "grandchild"}))

	ls := list.NewModel()
	ls.AddItems(list.MakeStringerList([]string{"second", "grandchild"}))

	firstLeave := boxer.NewLeave()
	firstLeave.Content = lf

	secondLeave := boxer.NewLeave()
	secondLeave.Content = ls

	rightChild := boxer.Model{}
	rightChild.Stacked = true
	rightChild.AddChildren([]boxer.BoxSize{{
		Box: boxer.Boxer(firstLeave),
	}, {
		Box: boxer.Boxer(secondLeave),
	}})

	leftList := list.NewModel()
	leftList.AddItems(list.MakeStringerList([]string{"leftList", "rootchild"}))

	leftLeave := boxer.NewLeave()
	leftLeave.Content = leftList
	leftLeave.Focus = true

	root := boxer.Model{}

	root.AddChildren([]boxer.BoxSize{
		{Box: leftLeave},
		{Box: rightChild},
	})
	p := tea.NewProgram(root)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}
