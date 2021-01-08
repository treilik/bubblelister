package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/boxer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	leftList := list.NewModel()
	leftList.AddItems(list.MakeStringerList([]string{"leftList", "rootchild"}))

	leftLeave := boxer.Model{}
	leftLeave.Border = true
	leftLeave.Childs = []boxer.BoxerSize{{Boxer: leftList}}

	lf := list.NewModel()
	lf.AddItems(list.MakeStringerList([]string{"first", "grandchild"}))

	ls := list.NewModel()
	ls.AddItems(list.MakeStringerList([]string{"second", "grandchild"}))

	firstLeave := boxer.Model{}
	firstLeave.Border = true
	firstLeave.Childs = []boxer.BoxerSize{{Boxer: lf}}

	secondLeave := boxer.Model{}
	secondLeave.Border = true
	secondLeave.Childs = []boxer.BoxerSize{{Boxer: ls}}

	rightChild := boxer.Model{}
	rightChild.Stacked = true
	rightChild.Childs = []boxer.BoxerSize{{
		Boxer: boxer.Boxer(firstLeave),
	}, {
		Boxer: boxer.Boxer(secondLeave),
	}}
	root := boxer.Model{}

	root.Childs = []boxer.BoxerSize{
		{Boxer: leftLeave},
		{Boxer: rightChild},
	}
	p := tea.NewProgram(root)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}
