package main

import (
	"io/ioutil"
	"log"
	"os"
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

	// start console gui
	initConsole()
}

func initConsole() error {
	g, err := guiInit()
	if err != nil {
		panic(err)
	}
	defer g.Close()

	return guiMainLoop(g)
}
