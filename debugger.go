package main

import "github.com/deweerdt/gocui"

func debuggerSetKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", 'n', gocui.ModNone, debuggerStep); err != nil {
		return err
	}

	return nil
}

func debuggerStep(g *gocui.Gui, v *gocui.View) error {
	Step()
	return nil
}
