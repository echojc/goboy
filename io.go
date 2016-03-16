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

	STAT_HBLANK uint8 = 0x08
	STAT_VBLANK uint8 = 0x10
	STAT_OAM    uint8 = 0x20
	STAT_LYC    uint8 = 0x40

	STAT_MODE_HBLANK uint8 = 0x00
	STAT_MODE_VBLANK uint8 = 0x01
	STAT_MODE_OAM    uint8 = 0x02
	STAT_MODE_LCD    uint8 = 0x03
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
		return STAT_MODE_VBLANK
	} else {
		x := cyclesWrapped - (cyclesPerLine * uint32(y))
		switch {
		case x < 80: // oam
			return STAT_MODE_OAM
		case x < 252: // oam + vram
			return STAT_MODE_LCD
		default: // hblank
			return STAT_MODE_HBLANK
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
