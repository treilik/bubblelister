package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/boxer"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	lowestFirst := list.NewModel()
	lowestFirst.AddItems(list.MakeStringerList([]string{"first", "lowest first child"}))
	lowestFirstLeave := boxer.NewLeave()
	lowestFirstLeave.Content = lowestFirst

	lowestSecond := list.NewModel()
	lowestSecond.AddItems(list.MakeStringerList([]string{"first", "lowest second child"}))
	lowestSecondLeave := boxer.NewLeave()
	lowestSecondLeave.Content = lowestSecond

	grandNode := boxer.Model{}
	grandNode.Stacked = true
	grandNode.AddChildren([]boxer.BoxSize{{Box: boxer.Boxer(lowestFirstLeave)}, {Box: boxer.Boxer(lowestSecondLeave)}})

	grandChild := list.NewModel()
	grandChild.AddItems(list.MakeStringerList([]string{"second", "grandchild"}))
	grandLeave := boxer.NewLeave()
	grandLeave.Content = grandChild
	grandLeave.Focus = true

	rightChild := boxer.Model{}
	rightChild.Stacked = true
	rightChild.AddChildren([]boxer.BoxSize{{
		Box: boxer.Boxer(grandNode),
	}, {
		Box: boxer.Boxer(grandLeave),
	}})

	leftList := list.NewModel()
	leftList.AddItems(list.MakeStringerList([]string{"leftList", "rootchild"}))
	leftLeave := boxer.NewLeave()
	leftLeave.Content = leftList

	rigthList := list.NewModel()
	rigthList.AddItems(list.MakeStringerList([]string{"rigthList", "rootchild"}))
	rigthLeave := boxer.NewLeave()
	rigthLeave.Content = rigthList

	root := boxer.Model{}
	root.AddChildren([]boxer.BoxSize{
		{Box: leftLeave},
		{Box: rigthLeave},
		{Box: rightChild},
	})
	p := tea.NewProgram(root)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}
