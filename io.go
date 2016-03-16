package main

const (
	REG_KEY  uint16 = 0xff00
	REG_IF   uint16 = 0xff0f
	REG_LCDC uint16 = 0xff40
	REG_STAT uint16 = 0xff41
	REG_LY   uint16 = 0xff44
	REG_LYC  uint16 = 0xff45
	REG_BGP  uint16 = 0xff47
	REG_OBP1 uint16 = 0xff48
	REG_OBP2 uint16 = 0xff49
	REG_IE   uint16 = 0xffff
)

func ioP1() uint8 {
	// TODO keys (0 = down, 1 = up)
	return 0x0f
}

func ioLy() uint8 {
	return uint8(cyclesWrapped / cyclesPerLine)
}

func ioLcdMode() uint8 {
	y := ioLy()
	if y >= 144 { // vblank
		return 1
	} else {
		x := cyclesWrapped - (cyclesPerLine * uint32(y))
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

func ioLcdYCoincidence() uint8 {
	if read(REG_LYC) == ioLy() {
		return 0x04
	} else {
		return 0
	}
}
