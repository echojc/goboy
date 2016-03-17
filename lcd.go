package main

import "sort"

type SpriteAttr uint16

type SpriteData struct {
	y        uint8
	x        uint8
	tile     uint8
	priority bool
	yFlip    bool
	xFlip    bool
	palette  bool
}

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

	SPRITE_WIDTH    = 8
	SPRITE_Y_OFFSET = 16
	SPRITE_X_OFFSET = 8

	LCD_WIDTH     = 160
	LCD_HEIGHT    = 144
	LCD_BG_WIDTH  = 256
	LCD_BG_HEIGHT = 256
)

// should always be LCD_WIDTH * LCD_HEIGHT
var buffer []uint8 = make([]uint8, LCD_WIDTH*LCD_HEIGHT)

func LcdSwapBuffers() {
	glSetBuffer(buffer)
	buffer = make([]uint8, LCD_WIDTH*LCD_HEIGHT)
}

func LcdBlitRow() {
	lcdc := read(REG_LCDC)
	scx := read(REG_SCX)
	scy := read(REG_SCY)
	y := ioLy() + scy
	bufferBaseOffset := int(y) * LCD_WIDTH

	// render bg
	if isBitSet(lcdc, LCDC_BG_WINDOW_ENABLE) {
		mapBaseAddr := GPU_TILE_MAP_0
		if isBitSet(lcdc, LCDC_BG_TILE_MAP) {
			mapBaseAddr = GPU_TILE_MAP_1
		}
		dataBaseAddr := GPU_TILE_DATA_0
		if isBitSet(lcdc, LCDC_BG_WINDOW_TILE_DATA) {
			dataBaseAddr = GPU_TILE_DATA_1
		}

		bgPalette := currentBgWindowPalette()

		for xOffset := 0; xOffset < LCD_WIDTH; xOffset++ {
			x := (xOffset + int(scx)) % LCD_BG_WIDTH
			tileDataAddr := tileDataAddr(mapBaseAddr, dataBaseAddr, uint8(x), y)

			tileX := uint8(x % 8)
			tileY := uint8(y % 8)
			tilePaletteIndex := tilePaletteIndex(tileDataAddr, tileX, tileY)

			buffer[bufferBaseOffset+xOffset] = bgPalette[tilePaletteIndex]
		}
	}

	// render sprites on top
	// TODO optimise with bg?
	if isBitSet(lcdc, LCDC_OBJ_ENABLE) {
		// TODO 8x16 sprites

		is8x16Mode := isBitSet(lcdc, LCDC_OBJ_8X16)
		sortedSprites := readSprites(y, is8x16Mode)
		xOffset := 0

		var spriteHeight uint8 = 8
		if is8x16Mode {
			spriteHeight = 16
		}

		for _, sprite := range sortedSprites {
			i := 0
			actualX := int(sprite.x) - SPRITE_X_OFFSET
			actualY := int(sprite.y) - SPRITE_Y_OFFSET
			objPalette := currentObjPalette(sprite.palette)

			if actualX < xOffset {
				i = xOffset - actualX
			}
			xOffset = actualX

			pixelY := uint8(int(y) - actualY)
			if sprite.yFlip {
				pixelY = (spriteHeight - 1) - pixelY
			}

			tileDataAddr := GPU_TILE_DATA_1 + (uint16(sprite.tile) << 4)
			for ; i < SPRITE_WIDTH; i++ {
				pixelX := uint8(i)
				if sprite.xFlip {
					pixelX = (SPRITE_WIDTH - 1) - pixelX
				}

				// 0 is transparent
				paletteIndex := tilePaletteIndex(tileDataAddr, pixelX, pixelY)
				if paletteIndex != 0 {
					buffer[bufferBaseOffset+xOffset] = objPalette[paletteIndex]
				}

				xOffset++
			}

			if xOffset >= LCD_WIDTH {
				break
			}
		}
	}
}

func tileDataAddr(mapBaseAddr, dataBaseAddr uint16, xPixel, yPixel uint8) uint16 {
	tileIdAddr := mapBaseAddr + (uint16(yPixel/8) * 0x20) + uint16(xPixel/8)
	tileId := read(tileIdAddr)
	tileDataAddr := dataBaseAddr + (uint16(tileId) << 4)

	if tileDataAddr >= 0x9800 {
		tileDataAddr -= 0x1000
	}

	return tileDataAddr
}

func tilePaletteIndex(baseAddr uint16, xPixel, yPixel uint8) uint8 {
	offsetAddr := baseAddr + uint16(yPixel)*2
	indexLo := (read(offsetAddr) >> (7 - xPixel)) & 0x01
	indexHi := (read(offsetAddr+1) >> (7 - xPixel)) & 0x01
	return (indexHi << 1) | indexLo
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

func currentObjPalette(isPalette1 bool) [4]uint8 {
	colors := [4]uint8{0xff, 0xaa, 0x55, 0x00}

	var palette uint8
	if isPalette1 {
		palette = read(REG_OBP1)
	} else {
		palette = read(REG_OBP0)
	}

	return [4]uint8{
		colors[palette&0x03],
		colors[(palette>>2)&0x03],
		colors[(palette>>4)&0x03],
		colors[(palette>>6)&0x03],
	}
}

func readSprites(y uint8, is8x16Mode bool) []*SpriteData {
	sprites := make([]*SpriteData, 0, 40)

	var spriteHeight uint8 = 8
	if is8x16Mode {
		spriteHeight = 16
	}

	var i uint16
	for i = 0; i < 40; i++ {
		sprite := readSprite(i)

		if sprite.y == 0 || y < sprite.y-SPRITE_Y_OFFSET || y >= sprite.y-SPRITE_Y_OFFSET+spriteHeight {
			continue
		}

		if sprite.x >= SPRITE_X_OFFSET+LCD_WIDTH {
			continue
		}

		sprites = append(sprites, sprite)
	}

	sort.Sort(SortableSpriteData(sprites))
	return sprites
}

func readSprite(id uint16) *SpriteData {
	if id >= 40 {
		return nil
	}

	offset := GPU_OAM + (4 * id)
	flags := read(offset + 3)

	return &SpriteData{
		read(offset),
		read(offset + 1),
		read(offset + 2),
		(flags>>7)&1 > 0,
		(flags>>6)&1 > 0,
		(flags>>5)&1 > 0,
		(flags>>4)&1 > 0,
	}
}

type SortableSpriteData []*SpriteData

func (s SortableSpriteData) Len() int {
	return len(s)
}

func (s SortableSpriteData) Less(i, j int) bool {
	if s[i].x != s[j].x {
		return s[i].x < s[j].x
	} else {
		return i < j
	}
}

func (s SortableSpriteData) Swap(i, j int) {
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}
