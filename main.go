package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/deweerdt/gocui"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const EnableGl bool = true

func main() {

	// read some data off disk
	if len(os.Args) < 1 {
		log.Println("no input file")
	} else {
		if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
			panic(err)
		} else {
			LoadRom(data)
		}
	}

	// init gocui
	g, err := guiInit()
	if err != nil {
		panic(err)
	}
	defer g.Close()

	if EnableGl {
		// init gl
		if err := glfw.Init(); err != nil {
			panic(err)
		}
		defer glfw.Terminate()

		if err := gl.Init(); err != nil {
			panic(err)
		}

		window, err := glCreateWindow()
		if err != nil {
			panic(err)
		}

		consoleErr := make(chan error, 1)
		go func() {
			consoleErr <- g.MainLoop()
		}()

		glMainLoop(window, g)

		if err := <-consoleErr; err != nil && err != gocui.ErrQuit {
			panic(err)
		}
	} else {
		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			panic(err)
		}
	}
}
