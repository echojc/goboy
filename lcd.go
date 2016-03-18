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
	y := ioLy()
	bgY := uint8((int(y) + int(scy)) % LCD_BG_HEIGHT)
	bufferBaseOffset := int(y) * LCD_WIDTH

	// palettes
	bgPalette := currentBgWindowPalette()
	obj0Palette := currentObjPalette(false)
	obj1Palette := currentObjPalette(true)

	// sprite height
	is8x16Mode := isBitSet(lcdc, LCDC_OBJ_8X16)
	var spriteHeight uint8 = 8
	if is8x16Mode {
		spriteHeight = 16
	}

	// get vram addresses
	bgMapBaseAddr := GPU_TILE_MAP_0
	if isBitSet(lcdc, LCDC_BG_TILE_MAP) {
		bgMapBaseAddr = GPU_TILE_MAP_1
	}
	//windowMapBaseAddr := GPU_TILE_MAP_0
	//if isBitSet(lcdc, LCDC_WINDOW_TILE_MAP) {
	//	windowMapBaseAddr = GPU_TILE_MAP_1
	//}
	bgWindowDataBaseAddr := GPU_TILE_DATA_0
	if isBitSet(lcdc, LCDC_BG_WINDOW_TILE_DATA) {
		bgWindowDataBaseAddr = GPU_TILE_DATA_1
	}

	// read sprites if necessary
	var sortedSprites []*SpriteData = nil
	if isBitSet(lcdc, LCDC_OBJ_ENABLE) {
		sortedSprites = readSprites(y, is8x16Mode)
	}

	spriteIndex := 0
	var x uint8
	for x = 0; x < LCD_WIDTH; x++ {

		// start with sprites
		var sprite *SpriteData = nil
		if isBitSet(lcdc, LCDC_OBJ_ENABLE) {

			// find first sprite that we should draw
			for ; spriteIndex < len(sortedSprites); spriteIndex++ {
				// cache the next sprite
				sprite = sortedSprites[spriteIndex]

				// not yet up to it for rendering, reset
				if x+SPRITE_X_OFFSET < sprite.x {
					sprite = nil
					break
				}

				// up to the sprite now, this is the sprite we want
				if x+SPRITE_X_OFFSET-SPRITE_WIDTH < sprite.x {
					break
				}
			}

			// cache sprite is nil for priority check later
			if spriteIndex == len(sortedSprites) {
				sprite = nil
			}

			// if we found a sprite to draw
			if sprite != nil {
				pixelX := x + SPRITE_X_OFFSET - sprite.x
				if sprite.xFlip {
					pixelX = (SPRITE_WIDTH - 1) - pixelX
				}

				pixelY := y + SPRITE_Y_OFFSET - sprite.y
				if sprite.yFlip {
					pixelY = (spriteHeight - 1) - pixelY
				}

				spriteTileAddr := spriteTileAddr(sprite.tile)
				paletteIndex := tilePaletteIndex(spriteTileAddr, pixelX, pixelY)

				// 0 is transparent
				if paletteIndex != 0 {
					objPalette := obj0Palette
					if sprite.palette {
						objPalette = obj1Palette
					}
					buffer[bufferBaseOffset+int(x)] = objPalette[paletteIndex]
					continue
				}
			}
		}

		// TODO draw window

		// draw bg
		if isBitSet(lcdc, LCDC_BG_WINDOW_ENABLE) {
			bgX := uint8((int(x) + int(scx)) % LCD_BG_WIDTH)
			bgWindowTileDataAddr := bgWindowTileDataAddr(bgMapBaseAddr, bgWindowDataBaseAddr, bgX, bgY)

			pixelX := uint8(bgX % 8)
			pixelY := uint8(bgY % 8)
			tilePaletteIndex := tilePaletteIndex(bgWindowTileDataAddr, pixelX, pixelY)

			buffer[bufferBaseOffset+int(x)] = bgPalette[tilePaletteIndex]
			continue
		}
	}
}

func bgWindowOffsetFromPixels(x, y uint8) uint16 {
	yOffset := uint16(y) / 8
	xOffset := uint16(x) / 8
	return yOffset*32 + xOffset
}

func bgWindowTileDataAddr(mapBaseAddr, dataBaseAddr uint16, xPixel, yPixel uint8) uint16 {
	tileIdAddr := mapBaseAddr + bgWindowOffsetFromPixels(xPixel, yPixel)
	tileId := read(tileIdAddr)
	tileDataAddr := dataBaseAddr + (uint16(tileId) << 4)

	if tileDataAddr >= 0x9800 {
		tileDataAddr -= 0x1000
	}

	return tileDataAddr
}

func spriteTileAddr(tileId uint8) uint16 {
	return GPU_TILE_DATA_1 + (uint16(tileId) << 4)
}

func tilePaletteIndex(baseAddr uint16, xPixel, yPixel uint8) uint8 {
	offsetAddr := baseAddr + uint16(yPixel)*2
	indexLo := (read(offsetAddr) >> (7 - xPixel)) & 0x01
	indexHi := (read(offsetAddr+1) >> (7 - xPixel)) & 0x01
	return (indexHi << 1) | indexLo
}

var paletteColors = [4]uint8{0xff, 0xaa, 0x55, 0x00}

func currentBgWindowPalette() [4]uint8 {
	palette := read(REG_BGP)
	return [4]uint8{
		paletteColors[palette&0x03],
		paletteColors[(palette>>2)&0x03],
		paletteColors[(palette>>4)&0x03],
		paletteColors[(palette>>6)&0x03],
	}
}

func currentObjPalette(isPalette1 bool) [4]uint8 {
	var palette uint8
	if isPalette1 {
		palette = read(REG_OBP1)
	} else {
		palette = read(REG_OBP0)
	}

	return [4]uint8{
		paletteColors[palette&0x03],
		paletteColors[(palette>>2)&0x03],
		paletteColors[(palette>>4)&0x03],
		paletteColors[(palette>>6)&0x03],
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

		if sprite.y == 0 || y+SPRITE_Y_OFFSET < sprite.y || y+SPRITE_Y_OFFSET-spriteHeight >= sprite.y {
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
