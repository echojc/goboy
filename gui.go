package main

import (
	"fmt"

	"github.com/deweerdt/gocui"
)

//////////////////////////////////////////
// Render entry point
//////////////////////////////////////////

func guiLayout(g *gocui.Gui) error {
	if err := updateRegistersView(g); err != nil {
		return err
	}

	if err := updateMiscView(g); err != nil {
		return err
	}

	if err := updateDisassemblerView(g); err != nil {
		return err
	}

	if err := updateMemoryView(g); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////
// Fixed-size views
//////////////////////////////////////////

const RegistersViewWidth = 30
const RegistersViewHeight = 7
const DisassemblerViewWidth = 48

//////////////////////////////////////////
// Registers
//////////////////////////////////////////

func updateRegistersView(g *gocui.Gui) error {
	v, err := g.SetView("registers", 0, 0, RegistersViewWidth, RegistersViewHeight)
	if err == gocui.ErrUnknownView {
		v.Title = " registers "
	} else if err != nil {
		return err
	}
	v.Clear()

	fmtString :=
		" A %08b %02x Z%d N%d H%d C%d\n" +
			" B %08b %02x %02x %08b C\n" +
			" D %08b %02x %02x %08b E\n" +
			" H %08b %02x %02x %08b L\n" +
			" SP %08b %08b %04x \n" +
			" PC %08b %08b %04x \n"

	fmt.Fprintf(v, fmtString,
		a, a, b2i(fz), b2i(fn), b2i(fh), b2i(fc),
		b, b, c, c,
		d, d, e, e,
		h, h, l, l,
		sp>>8, sp&0xff, sp,
		pc>>8, pc&0xff, pc)

	return nil
}

func b2i(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

//////////////////////////////////////////
// Disassembler
//////////////////////////////////////////

func updateDisassemblerView(g *gocui.Gui) error {
	_, maxY := g.Size()
	height := (maxY - 1) - RegistersViewHeight

	v, err := g.SetView("disassembler", 0, RegistersViewHeight, DisassemblerViewWidth, maxY-1)
	if err == gocui.ErrUnknownView {
		v.Title = " disassembler "
	} else if err != nil {
		return err
	}
	v.Clear()

	// start from up to `extraLines` instructions ago
	extraLines := uint16(height / 2)
	maxExtraBytes := extraLines * 3

	var startAddr uint16
	if pc < maxExtraBytes {
		startAddr = 0
	} else {
		startAddr = pc - maxExtraBytes
	}

	// get the entire range
	ds := renderDisassembly(startAddr, height+int(maxExtraBytes))

	// find where instruction actually starts, and count `extraLines` back
	var startIndex int
	for i := 0; i < len(ds); i++ {
		if ds[i].addr == pc {
			startIndex = i - int(extraLines)
			break
		}
	}
	if startIndex < 0 {
		startIndex = 0
	}

	// print the disassembly
	for i := startIndex; i < startIndex+height && i < len(ds); i++ {
		fmt.Fprintf(v, " %s\n", ds[i].pretty)
	}

	v.SetCursor(0, height/2)
	v.Highlight = true

	return nil
}

type Disassembly struct {
	addr   uint16
	pretty string
}

func renderDisassembly(startAddr uint16, max int) []Disassembly {
	// allocate enough memory to hold all instructions
	output := make([]Disassembly, max)

	// render all instructions
	addr := uint(startAddr)
	for i := 0; i < max && addr < 0x10000; i++ {
		raw, pretty, length := Disassemble(uint16(addr))
		output[i] = Disassembly{uint16(addr), fmt.Sprintf("%04x    % -12s%s", addr, raw, pretty)}
		addr += length
	}

	return output
}

//////////////////////////////////////////
// Memory
//////////////////////////////////////////

var viewMemoryBaseAddr uint16

func updateMemoryView(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	height := (maxY - 1) - RegistersViewHeight

	v, err := g.SetView("memory", DisassemblerViewWidth, RegistersViewHeight, maxX-1, maxY-1)
	if err == gocui.ErrUnknownView {
		v.Title = " memory "
	} else if err != nil {
		return err
	}
	v.Clear()

	// header
	fmt.Fprintf(v, "         00 01 02 03 04 05 06 07   08 09 0a 0b 0c 0d 0e 0f\n\n")

	// start from nearest 0x10 rounded down
	addr := uint(viewMemoryBaseAddr - (viewMemoryBaseAddr % 0x10))
	for i := 0; i < height && addr < 0x10000; i++ {
		fmt.Fprintf(v, " %04x   ", addr)
		for j := uint16(0x00); j < 0x08; j++ {
			fmt.Fprintf(v, " %02x", read(uint16(addr)+j))
		}
		fmt.Fprintf(v, "  ")
		for j := uint16(0x08); j < 0x10; j++ {
			fmt.Fprintf(v, " %02x", read(uint16(addr)+j))
		}
		fmt.Fprintf(v, "\n")
		addr += 0x10
	}

	return nil
}

//////////////////////////////////////////
// Misc
//////////////////////////////////////////

func updateMiscView(g *gocui.Gui) error {
	maxX, _ := g.Size()
	v, err := g.SetView("misc", RegistersViewWidth, 0, maxX-1, RegistersViewHeight)
	if err == gocui.ErrUnknownView {
	} else if err != nil {
		return err
	}
	v.Clear()

	return nil
}
