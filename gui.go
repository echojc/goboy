package main

import (
	"fmt"
	"sort"

	"github.com/deweerdt/gocui"
)

var guiCompleted bool = false

func guiInit() (*gocui.Gui, error) {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		return nil, err
	}

	g.SetLayout(guiLayout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, guiQuit); err != nil {
		return nil, err
	}

	if err := guiSetKeybindings(g); err != nil {
		return nil, err
	}

	return g, nil
}

func guiQuit(g *gocui.Gui, v *gocui.View) error {
	guiCompleted = true
	return gocui.ErrQuit
}

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
const IoViewWidth = 21

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
	fmt.Fprintf(v, " IE   %08b %02x\n", read(REG_IE), read(REG_IE))
	fmt.Fprintf(v, " IF   %08b %02x\n", read(REG_IF), read(REG_IF))

	fmt.Fprintln(v)
	fmt.Fprintf(v, "      M W   OB\n")
	fmt.Fprintf(v, " LCDC %08b %02x\n", read(REG_LCDC), read(REG_LCDC))

	fmt.Fprintln(v)
	fmt.Fprintf(v, "       YOVHC M\n")
	fmt.Fprintf(v, " STAT %08b %02x\n", read(REG_STAT), read(REG_STAT))
	fmt.Fprintf(v, " LY   %08b %02x\n", read(REG_LY), read(REG_LY))
	fmt.Fprintf(v, " LYC  %08b %02x\n", read(REG_LYC), read(REG_LYC))

	fmt.Fprintln(v)
	fmt.Fprintf(v, " BGP  %08b %02x\n", read(REG_BGP), read(REG_BGP))
	fmt.Fprintf(v, " OBP0 %08b %02x\n", read(REG_OBP0), read(REG_OBP0))
	fmt.Fprintf(v, " OBP1 %08b %02x\n", read(REG_OBP1), read(REG_OBP1))

	fmt.Fprintln(v)
	fmt.Fprintf(v, " SCX  %08b %02x\n", read(REG_SCX), read(REG_SCX))
	fmt.Fprintf(v, " SCY  %08b %02x\n", read(REG_SCY), read(REG_SCY))
	fmt.Fprintf(v, " WX   %08b %02x\n", read(REG_WX), read(REG_WX))
	fmt.Fprintf(v, " WY   %08b %02x\n", read(REG_WY), read(REG_WY))

	fmt.Fprintln(v)
	fmt.Fprintf(v, " DIV  %08b %02x\n", read(REG_DIV), read(REG_DIV))
	fmt.Fprintf(v, " TIMA %08b %02x\n", read(REG_TIMA), read(REG_TIMA))
	fmt.Fprintf(v, " TMA  %08b %02x\n", read(REG_TMA), read(REG_TMA))
	fmt.Fprintf(v, " TAC  %08b %02x\n", read(REG_TAC), read(REG_TAC))

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

	if debuggerRunning {
		fmt.Fprintln(v, " running")
	} else {
		fmt.Fprintln(v, " paused")
	}

	fmt.Fprintf(v, " cycles: %d", cyclesWrapped)
	if halted {
		fmt.Fprintln(v, " (halted)")
	}

	return nil
}

//////////////////////////////////////////
// Keybindings
//////////////////////////////////////////

func guiSetKeybindings(g *gocui.Gui) error {
	// debugging
	if err := g.SetKeybinding("", 'n', gocui.ModNone, actionGui(debuggerStep)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'r', gocui.ModNone, actionGui(debuggerRun)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlB, gocui.ModNone, actionGui(debuggerBreak)); err != nil {
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

	// exports
	if err := g.SetKeybinding("", gocui.KeyCtrlT, gocui.ModNone, action(ExportTileData)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlM, gocui.ModNone, action(ExportTileMap0)); err != nil {
		return err
	}

	// interrupts
	g.SetKeybinding("", gocui.KeyCtrlV, gocui.ModNone, action(func() { triggerInterrupt(INT_VBLANK) }))

	return nil
}

func action(fn func()) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		fn()
		return nil
	}
}

func actionGui(fn func(g *gocui.Gui)) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		fn(g)
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
