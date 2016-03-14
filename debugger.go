package main

import "time"

var debuggerBreakpoints []uint16 = []uint16{}
var debuggerEvents chan DebuggerEvent

type DebuggerEvent int

const (
	DEBUGGER_RUN   DebuggerEvent = 0
	DEBUGGER_STEP  DebuggerEvent = 1
	DEBUGGER_BREAK DebuggerEvent = 2
	DEBUGGER_QUIT  DebuggerEvent = 3
)

// 2^22 / 10^9
const cyclesPerNanosecond = 0.004194304

func debuggerLoop(events <-chan DebuggerEvent) {
	running := false
	ticker := time.Tick(time.Millisecond)
	var startCycle uint64
	var startTime int64

	for {
		select {
		case ev := <-events:
			switch ev {
			case DEBUGGER_RUN:
				running = true
				viewDisassemblerPcLock = false
				startCycle = cycles
				startTime = time.Now().UnixNano()
			case DEBUGGER_STEP:
				if !running {
					viewDisassemblerPcLock = true
					Step()
				}
			case DEBUGGER_BREAK:
				running = false
				viewDisassemblerPcLock = true
			case DEBUGGER_QUIT:
				return
			}
		case <-ticker:
			if running {
				for {
					// check for breakpoints
					if isBreakpoint(pc) {
						break
					}

					// if we're ahead of where we're meant to be, take a break
					now := time.Now().UnixNano()
					expectedCycles := uint64(float64(now-startTime) * cyclesPerNanosecond)
					if cycles > startCycle+expectedCycles {
						break
					}

					// go for it
					Step()
				}
			}
		}
	}
}

func debuggerInit() {
	debuggerEvents = make(chan DebuggerEvent, 1000)
	go debuggerLoop(debuggerEvents)
}

func debuggerStep() {
	debuggerEvents <- DEBUGGER_STEP
}

func debuggerRun() {
	debuggerEvents <- DEBUGGER_RUN
}

func debuggerBreak() {
	debuggerEvents <- DEBUGGER_BREAK
}

func debuggerQuit() {
	debuggerEvents <- DEBUGGER_QUIT
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
