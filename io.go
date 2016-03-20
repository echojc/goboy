package main

const (
	REG_KEY  uint16 = 0xff00
	REG_DIV  uint16 = 0xff04
	REG_TIMA uint16 = 0xff05
	REG_TMA  uint16 = 0xff06
	REG_TAC  uint16 = 0xff07
	REG_IF   uint16 = 0xff0f
	REG_LCDC uint16 = 0xff40
	REG_STAT uint16 = 0xff41
	REG_SCY  uint16 = 0xff42
	REG_SCX  uint16 = 0xff43
	REG_LY   uint16 = 0xff44
	REG_LYC  uint16 = 0xff45
	REG_DMA  uint16 = 0xff46
	REG_BGP  uint16 = 0xff47
	REG_OBP0 uint16 = 0xff48
	REG_OBP1 uint16 = 0xff49
	REG_WY   uint16 = 0xff4a
	REG_WX   uint16 = 0xff4b
	REG_IE   uint16 = 0xffff

	LCDC_ENABLE              uint8 = 0x80
	LCDC_WINDOW_TILE_MAP     uint8 = 0x40 // 0 = 0x9800, 1 = 0x9c00
	LCDC_WINDOW_ENABLE       uint8 = 0x20
	LCDC_BG_WINDOW_TILE_DATA uint8 = 0x10 // 0 = 0x9000, 1 = 0x8000
	LCDC_BG_TILE_MAP         uint8 = 0x08 // 0 = 0x9800, 1 = 0x9c00
	LCDC_OBJ_8X16            uint8 = 0x04 // 0 = 8x8, 1 = 8x16
	LCDC_OBJ_ENABLE          uint8 = 0x02
	LCDC_BG_WINDOW_ENABLE    uint8 = 0x01

	STAT_LYC    uint8 = 0x40
	STAT_OAM    uint8 = 0x20
	STAT_VBLANK uint8 = 0x10
	STAT_HBLANK uint8 = 0x08

	STAT_MODE_HBLANK uint8 = 0x00
	STAT_MODE_VBLANK uint8 = 0x01
	STAT_MODE_OAM    uint8 = 0x02
	STAT_MODE_LCD    uint8 = 0x03

	TIMER_MODE_4096   uint8 = 0x00
	TIMER_MODE_262144 uint8 = 0x01
	TIMER_MODE_65536  uint8 = 0x02
	TIMER_MODE_16384  uint8 = 0x03
	TIMER_ENABLE      uint8 = 0x04

	KEY_DOWN   uint8 = 0x80
	KEY_UP     uint8 = 0x40
	KEY_LEFT   uint8 = 0x20
	KEY_RIGHT  uint8 = 0x10
	KEY_START  uint8 = 0x08
	KEY_SELECT uint8 = 0x04
	KEY_B      uint8 = 0x02
	KEY_A      uint8 = 0x01
)

func isBitSetAddr(addr uint16, bit uint8) bool {
	return (read(addr) & bit) > 0
}

func isBitSet(v uint8, bit uint8) bool {
	return (v & bit) > 0
}

// keys are backwards: 0 = down, 1 = up
var ioKeys uint8 = 0xff

func ioKeyDown(key uint8) {
	ioKeys &= ^key
	z80KeyDown = true
}

func ioKeyUp(key uint8) {
	ioKeys |= key
}

func ioP1() uint8 {
	keys := io[0] // can't use read() because recursive

	if (keys & 0x10) == 0 { // bit 4 is low == dpad
		return ioKeys >> 4
	} else if (keys & 0x20) == 0 { // bit 5 is low == buttons
		return ioKeys & 0x0f
	} else {
		return 0x0f
	}
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
