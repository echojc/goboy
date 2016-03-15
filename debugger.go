package main

import (
	"time"

	"github.com/deweerdt/gocui"
)

var debuggerBreakpoints []uint16 = []uint16{}
var debuggerEvents chan *DebuggerEvent

type EventEnum int

const (
	DEBUGGER_RUN   EventEnum = 0
	DEBUGGER_STEP  EventEnum = 1
	DEBUGGER_BREAK EventEnum = 2
	DEBUGGER_QUIT  EventEnum = 3
)

type DebuggerEvent struct {
	G     *gocui.Gui
	Event EventEnum
}

// 2^22 / 10^9
const cyclesPerNanosecond = 0.004194304

func debuggerLoop(events <-chan *DebuggerEvent) {
	running := false
	ticker := time.Tick(time.Millisecond)
	var startCycle uint64
	var startTime int64
	var runGui *gocui.Gui

	for {
		select {
		case ev := <-events:
			switch ev.Event {
			case DEBUGGER_RUN:
				running = true
				startCycle = cycles
				startTime = time.Now().UnixNano()
				runGui = ev.G
				viewDisassemblerPcLock = false
				debuggerUpdateGui(ev.G)
			case DEBUGGER_STEP:
				if !running {
					Step()
					viewDisassemblerPcLock = true
					debuggerUpdateGui(ev.G)
				}
			case DEBUGGER_BREAK:
				running = false
				viewDisassemblerPcLock = true
				debuggerUpdateGui(ev.G)
			case DEBUGGER_QUIT:
				return
			}
		case <-ticker:
			if running {
				for {
					// step first to avoid double breakpoint issues
					Step()

					// check for breakpoints - stop running if found
					if isBreakpoint(pc) {
						running = false
						viewDisassemblerPcLock = true
						debuggerUpdateGui(runGui)
						break
					}

					// if we're ahead of where we're meant to be, yield for now
					now := time.Now().UnixNano()
					expectedCycles := uint64(float64(now-startTime) * cyclesPerNanosecond)
					if cycles > startCycle+expectedCycles {
						break
					}
				}
			}
		}
	}
}

func debuggerInit() {
	debuggerEvents = make(chan *DebuggerEvent, 10)
	go debuggerLoop(debuggerEvents)
}

func debuggerStep(g *gocui.Gui) {
	debuggerEvents <- &DebuggerEvent{g, DEBUGGER_STEP}
}

func debuggerRun(g *gocui.Gui) {
	debuggerEvents <- &DebuggerEvent{g, DEBUGGER_RUN}
}

func debuggerBreak(g *gocui.Gui) {
	debuggerEvents <- &DebuggerEvent{g, DEBUGGER_BREAK}
}

func debuggerQuit() {
	debuggerEvents <- &DebuggerEvent{nil, DEBUGGER_QUIT}
}

func debuggerUpdateGui(g *gocui.Gui) {
	g.Execute(func(g *gocui.Gui) error { return nil })
}

func debuggerToggleBreakpoint(addr uint16) {
	for i := 0; i < len(debuggerBreakpoints); i++ {
		if debuggerBreakpoints[i] == addr {
			// delete
			debuggerBreakpoints = append(debuggerBreakpoints[:i], debuggerBreakpoints[i+1:]...)
			return
		}
	}
	// didn't find it, append
	debuggerBreakpoints = append(debuggerBreakpoints, addr)
}

func isBreakpoint(addr uint16) bool {
	for _, v := range debuggerBreakpoints {
		if addr == v {
			return true
		}
	}
	return false
}
