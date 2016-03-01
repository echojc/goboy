package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/deweerdt/gocui"
)

func main() {

	// read some data off disk
	if len(os.Args) < 1 {
		log.Println("no input file")
	} else {
		if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
			log.Fatal(err)
		} else {
			LoadRom(data)
		}
	}

	// start gui
	initGui()
}

func initGui() {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetLayout(guiLayout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := debuggerSetKeybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
