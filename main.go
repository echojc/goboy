package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	// read some data off disk
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// load data into some implemented part of memory
	var i uint16
	for i = 0x0000; i < 0x2000; i++ {
		write(i+0xc000, data[i])
	}

	// print the disassembly
	for _, line := range renderDisassembly(0xc000, 0xc400) {
		fmt.Println(line)
	}
}

func renderDisassembly(startAddr, endAddr uint16) []string {
	// allocate enough memory to hold all instructions
	output := make([]string, 0, endAddr-startAddr)

	// render all instructions
	for addr := startAddr; addr < endAddr; {
		raw, pretty, length := Disassemble(addr)
		output = append(output, fmt.Sprintf("0x%04x\t% -12s%s", addr, raw, pretty))
		addr += length
	}

	return output
}
