package main

type SpriteAttr uint16

const (
	GPU_TILE_MAP_0       uint16 = 0x9800
	GPU_TILE_MAP_1       uint16 = 0x9c00
	GPU_TILE_DATA_0      uint16 = 0x9000
	GPU_TILE_DATA_1      uint16 = 0x8000
	GPU_TILE_DATA_SHARED uint16 = 0x8800
	GPU_OAM              uint16 = 0xfe00

	SPRITE_Y        SpriteAttr = 0
	SPRITE_X        SpriteAttr = 1
	SPRITE_TILE     SpriteAttr = 2
	SPRITE_PALETTE  SpriteAttr = 3
	SPRITE_XFLIP    SpriteAttr = 5
	SPRITE_YFLIP    SpriteAttr = 6
	SPRITE_PRIORITY SpriteAttr = 7

	LCD_WIDTH     = 160
	LCD_HEIGHT    = 144
	LCD_BG_WIDTH  = 256
	LCD_BG_HEIGHT = 256
)

// should always be LCD_WIDTH * LCD_HEIGHT
var buffer []uint8

func LcdInit() {
	buffer = make([]uint8, LCD_WIDTH*LCD_HEIGHT)
}

func LcdSwapBuffers() {
	glSetBuffer(buffer)
	buffer = make([]uint8, LCD_WIDTH*LCD_HEIGHT)
}

func LcdBlitRow() {
	lcdc := read(REG_LCDC)
	scx := read(REG_SCX)
	scy := read(REG_SCY)
	y := ioLy() + scy

	// render bg
	tileMapBaseAddr := GPU_TILE_MAP_0
	if isBitSet(lcdc, LCDC_BG_TILE_MAP) {
		tileMapBaseAddr = GPU_TILE_MAP_1
	}
	tileDataBaseAddr := GPU_TILE_DATA_0
	if isBitSet(lcdc, LCDC_BG_WINDOW_TILE_DATA) {
		tileDataBaseAddr = GPU_TILE_DATA_1
	}

	bufferBaseOffset := int(y) * LCD_WIDTH
	bgPalette := currentBgWindowPalette()

	for xOffset := 0; xOffset < LCD_WIDTH; xOffset++ {
		x := (xOffset + int(scx)) % LCD_BG_WIDTH
		tileDataIndexAddr := tileMapBaseAddr + (uint16(y/8) * 0x20) + uint16(x/8)
		tileDataAddr := tileDataBaseAddr + uint16(read(tileDataIndexAddr))<<4
		if tileDataAddr > 0x9800 {
			tileDataAddr -= 0x1000
		}

		tileX := uint(x % 8)
		tileY := y % 8
		tileAddr := tileDataAddr + (uint16(tileY) * 2)
		tileLo := read(tileAddr)
		tileHi := read(tileAddr + 1)
		tilePaletteIndex := (((tileHi >> (7 - tileX)) & 0x01) << 1) | ((tileLo >> (7 - tileX)) & 0x01)

		buffer[bufferBaseOffset+xOffset] = bgPalette[tilePaletteIndex]
	}
}

func readSprite(id uint16, attr SpriteAttr) uint8 {
	if id >= 40 {
		return 0
	}

	switch attr {
	case SPRITE_Y:
		return read(GPU_OAM + (4 * id))
	case SPRITE_X:
		return read(GPU_OAM + (4 * id) + 1)
	case SPRITE_TILE:
		return read(GPU_OAM + (4 * id) + 2)
	case SPRITE_PALETTE:
		return (read(GPU_OAM+(4*id)+3) >> 4) & 0x01
	case SPRITE_XFLIP:
		return (read(GPU_OAM+(4*id)+3) >> 5) & 0x01
	case SPRITE_YFLIP:
		return (read(GPU_OAM+(4*id)+3) >> 6) & 0x01
	case SPRITE_PRIORITY:
		return (read(GPU_OAM+(4*id)+3) >> 7) & 0x01
	}

	return 0
}

func currentBgWindowPalette() [4]uint8 {
	colors := [4]uint8{0xff, 0xaa, 0x55, 0x00}
	palette := read(REG_BGP)
	return [4]uint8{
		colors[palette&0x03],
		colors[(palette>>2)&0x03],
		colors[(palette>>4)&0x03],
		colors[(palette>>6)&0x03],
	}
}
