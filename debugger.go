package main

var debuggerBreakpoints []uint16 = []uint16{}

func debuggerStep() {
	// snap to pc
	viewDisassemblerPcLock = true
	Step()
}

func debuggerRun() {
	viewDisassemblerPcLock = true
	Step()
	for !isBreakpoint(pc) {
		Step()
	}
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
