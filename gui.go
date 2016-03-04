package main

import (
	"fmt"
	"sort"

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

	if err := updateMemoryView(g); err != nil {
		return err
	}

	if err := updateDisassemblerView(g); err != nil {
		return err
	}

	if err := updateIoView(g); err != nil {
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
const IoViewWidth = 18

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

var viewDisassemblerBaseIndex int
var disassembly []Disassembly
var viewDisassemblerPcLock bool
var viewDisassemblerCursorIndex int

var refreshes int = 0

func updateDisassemblerView(g *gocui.Gui) error {
	_, maxY := g.Size()
	height := (maxY - 1) - RegistersViewHeight

	v, err := g.SetView("disassembler", 0, RegistersViewHeight, DisassemblerViewWidth, maxY-1)
	if err == gocui.ErrUnknownView {
		g.SetCurrentView(v.Name())

		// init
		disassembly = renderDisassembly(0)
		viewDisassemblerBaseIndex = indexOfDisassemblyForAddress(0x0100)
		viewDisassemblerPcLock = true
	} else if err != nil {
		return err
	}
	v.Clear()

	pcIndex := indexOfDisassemblyForAddress(pc)

	// if pc lock, scroll to pc
	if viewDisassemblerPcLock && (pcIndex < viewDisassemblerBaseIndex || pcIndex >= viewDisassemblerBaseIndex+height-1) {
		viewDisassemblerBaseIndex = pcIndex
	}

	// decide if we need to rerender disassembly
	if z80SmallestDirtyAddr < disassembly[viewDisassemblerBaseIndex+height-2].addr {
		splicePoint := sort.Search(len(disassembly), func(i int) bool {
			return disassembly[i].addr >= z80SmallestDirtyAddr
		})
		disassembly = append(disassembly[:splicePoint], renderDisassembly(z80SmallestDirtyAddr)...)
		z80SmallestDirtyAddr = 0xffff

		// redo all the pc stuff
		pcIndex = indexOfDisassemblyForAddress(pc)
		if viewDisassemblerPcLock && (pcIndex < viewDisassemblerBaseIndex || pcIndex >= viewDisassemblerBaseIndex+height-1) {
			viewDisassemblerBaseIndex = pcIndex
		}

		refreshes += 1
	}

	// update title to include rerender count
	if g.CurrentView() == v {
		v.Title = fmt.Sprintf(" disassembler* %d/%04x ", refreshes, z80SmallestDirtyAddr)
	} else {
		v.Title = fmt.Sprintf(" disassembler %d/%04x ", refreshes, z80SmallestDirtyAddr)
	}

	// put cursor at the current instruction
	v.SetCursor(0, viewDisassemblerCursorIndex)
	v.Highlight = true

	// print the disassembly
	fmtString := fmt.Sprintf("\033[%%dm%%c %% -%ds\n", DisassemblerViewWidth)
	for i := viewDisassemblerBaseIndex; i < viewDisassemblerBaseIndex+height && i < len(disassembly); i++ {
		// bg color for breakpoints/pc
		bgColor := 0
		switch {
		case pc == disassembly[i].addr:
			bgColor = 42 // bg green
		case isBreakpoint(disassembly[i].addr):
			bgColor = 41 // bg red
		}

		// additional indicator of current pc
		arrow := ' '
		if pc == disassembly[i].addr {
			arrow = '>'
		}

		fmt.Fprintf(v, fmtString, bgColor, arrow, disassembly[i].pretty)
	}

	return nil
}

func indexOfDisassemblyForAddress(addr uint16) int {
	return sort.Search(len(disassembly), func(i int) bool {
		return disassembly[i].addr >= addr
	})
}

//////////////////////////////////////////
// Memory
//////////////////////////////////////////

var viewMemoryBaseAddr int = 0x0000

func updateMemoryView(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	height := (maxY - 1) - RegistersViewHeight

	v, err := g.SetView("memory", DisassemblerViewWidth, RegistersViewHeight, maxX-IoViewWidth, maxY-1)
	if err == gocui.ErrUnknownView {
		g.SetCurrentView(v.Name())
	} else if err != nil {
		return err
	}
	v.Clear()

	if g.CurrentView() == v {
		v.Title = " memory* "
	} else {
		v.Title = " memory "
	}

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
// I/O registers
//////////////////////////////////////////

func updateIoView(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView("io", maxX-IoViewWidth, RegistersViewHeight, maxX-1, maxY-1)
	if err == gocui.ErrUnknownView {
		v.Title = " i/o "
	} else if err != nil {
		return err
	}
	v.Clear()

	var interruptStatus = "off"
	if interruptsEnabled {
		interruptStatus = "on "
	}
	fmt.Fprintf(v, " I:%s   KSTLV\n", interruptStatus)
	fmt.Fprintf(v, " IE   %08b\n", read(0xffff))
	fmt.Fprintf(v, " IF   %08b\n", read(0xff0f))

	fmt.Fprintf(v, "      M W   OB\n")
	fmt.Fprintf(v, " LCDC %08b\n", read(0xff40))

	fmt.Fprintf(v, "       YOVHC M\n")
	fmt.Fprintf(v, " STAT %08b\n", read(0xff41))
	fmt.Fprintf(v, " LY   %08b\n", read(0xff44))
	fmt.Fprintf(v, " LYC  %08b\n", read(0xff45))

	fmt.Fprintf(v, " BGP  %08b\n", read(0xff47))
	fmt.Fprintf(v, " OBP1 %08b\n", read(0xff48))
	fmt.Fprintf(v, " OBP2 %08b\n", read(0xff49))

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

	fmt.Fprintf(v, " clks: % 5d\n", cycles)

	return nil
}

//////////////////////////////////////////
// Keybindings
//////////////////////////////////////////

func guiSetKeybindings(g *gocui.Gui) error {
	// debugging
	if err := g.SetKeybinding("", 'n', gocui.ModNone, action(debuggerStep)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'r', gocui.ModNone, action(debuggerRun)); err != nil {
		return err
	}

	// pane
	if err := g.SetKeybinding("", 'd', gocui.ModNone, guiSetFocus("disassembler")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'm', gocui.ModNone, guiSetFocus("memory")); err != nil {
		return err
	}

	// scroll
	if err := g.SetKeybinding("", gocui.KeyCtrlE, gocui.ModNone, guiScroll("e")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlY, gocui.ModNone, guiScroll("y")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, guiScroll("d")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, guiScroll("u")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'j', gocui.ModNone, guiScroll("j")); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'k', gocui.ModNone, guiScroll("k")); err != nil {
		return err
	}

	// debugger
	if err := g.SetKeybinding("", 'b', gocui.ModNone, actionToggleBreakpoint()); err != nil {
		return err
	}

	return nil
}

func action(fn func()) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		fn()
		return nil
	}
}

func guiSetFocus(viewname string) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		return g.SetCurrentView(viewname)
	}
}

func guiScroll(key string) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		view := g.CurrentView()
		if view == nil {
			return nil
		}

		_, maxY := g.Size()
		height := (maxY - 1) - RegistersViewHeight

		switch view.Name() {
		case "disassembler":
			switch key {
			case "e":
				viewDisassemblerBaseIndex += 1
			case "y":
				viewDisassemblerBaseIndex -= 1
			case "d":
				viewDisassemblerBaseIndex += height / 2
			case "u":
				viewDisassemblerBaseIndex -= height / 2
			case "j":
				viewDisassemblerCursorIndex += 1
			case "k":
				viewDisassemblerCursorIndex -= 1
			}
			if viewDisassemblerBaseIndex < 0 {
				viewDisassemblerBaseIndex = 0
			}
			if viewDisassemblerBaseIndex > len(disassembly)-height+1 {
				viewDisassemblerBaseIndex = len(disassembly) - height + 1
			}
			if viewDisassemblerCursorIndex < 0 {
				viewDisassemblerCursorIndex = 0
			}
			if viewDisassemblerCursorIndex > height-1 {
				viewDisassemblerCursorIndex = height - 2
			}
			viewDisassemblerPcLock = false
		case "memory":
			switch key {
			case "e":
				viewMemoryBaseAddr += 0x10
			case "y":
				viewMemoryBaseAddr -= 0x10
			case "d":
				viewMemoryBaseAddr += 0x10 * height / 2
			case "u":
				viewMemoryBaseAddr -= 0x10 * height / 2
			}
			if viewMemoryBaseAddr < 0 {
				viewMemoryBaseAddr = 0
			}
			if viewMemoryBaseAddr > 0x10000-(0x10*(height-3)) {
				viewMemoryBaseAddr = 0x10000 - (0x10 * (height - 3))
			}
		}

		return nil
	}
}

func actionToggleBreakpoint() gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		addr := disassembly[viewDisassemblerBaseIndex+viewDisassemblerCursorIndex].addr
		debuggerToggleBreakpoint(addr)
		return nil
	}
}
