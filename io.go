package main

func ioP1() uint8 {
	// TODO keys (0 = down, 1 = up)
	return 0x0f
}

func ioLy() uint8 {
	return uint8(cycles / cyclesPerLine)
}

func ioLcdMode() uint8 {
	y := ioLy()
	if y >= 144 { // vblank
		return 1
	} else {
		x := cycles - (cyclesPerLine * uint32(y))
		switch {
		case x < 80: // oam
			return 2
		case x < 252: // oam + vram
			return 3
		default: // hblank
			return 0
		}
	}
}
