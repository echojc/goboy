package main

func debuggerStep() {
	// snap to pc
	viewDisassemblerPcLock = true
	Step()
}

func debuggerRun() {
	for pc != 0x01a1 {
		debuggerStep()
	}
}
