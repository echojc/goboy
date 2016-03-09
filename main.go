package main

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/deweerdt/gocui"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const EnableGl bool = true

func init() {
	runtime.LockOSThread()
}

func main() {

	// read some data off disk
	if len(os.Args) < 1 {
		log.Println("no input file")
		os.Exit(2)
	} else {
		if data, err := ioutil.ReadFile(os.Args[1]); err != nil {
			panic(err)
		} else {
			LoadRom(data)
		}
	}

	// set up logger
	logFile, err := os.OpenFile("goboy.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetFlags(log.LstdFlags)
	log.SetOutput(logFile)

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
