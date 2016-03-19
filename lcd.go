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
	pixels   [8]uint8
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

	WINDOW_X_OFFSET = 7

	LCD_WIDTH         = 160
	LCD_HEIGHT        = 144
	LCD_BG_WIDTH      = 256
	LCD_BG_HEIGHT     = 256
	LCD_WINDOW_WIDTH  = 256
	LCD_WINDOW_HEIGHT = 256

	PIXEL_BG           uint8 = 0x00
	PIXEL_BG_0         uint8 = 0x00
	PIXEL_BG_1         uint8 = 0x01
	PIXEL_BG_2         uint8 = 0x02
	PIXEL_BG_3         uint8 = 0x03
	PIXEL_WINDOW       uint8 = 0x04
	PIXEL_WINDOW_0     uint8 = 0x04
	PIXEL_WINDOW_1     uint8 = 0x05
	PIXEL_WINDOW_2     uint8 = 0x06
	PIXEL_WINDOW_3     uint8 = 0x07
	PIXEL_OBJ0         uint8 = 0x08
	PIXEL_OBJ0_1       uint8 = 0x09
	PIXEL_OBJ0_2       uint8 = 0x0a
	PIXEL_OBJ0_3       uint8 = 0x0b
	PIXEL_OBJ1         uint8 = 0x0c
	PIXEL_OBJ1_1       uint8 = 0x0d
	PIXEL_OBJ1_2       uint8 = 0x0e
	PIXEL_OBJ1_3       uint8 = 0x0f
	PIXEL_LOW_PRIORITY uint8 = 0x80
)

// should always be LCD_WIDTH * LCD_HEIGHT
var buffer []uint8

// REG_WY is read once per frame
var wy uint8

// windowY only increments if window is enabled
var windowY uint8

func LcdInit() {
	lcdResetBuffer()
}

func LcdSwapBuffers() {
	glSetBuffer(buffer)
	lcdResetBuffer()
}

func lcdResetBuffer() {
	buffer = make([]uint8, LCD_WIDTH*LCD_HEIGHT)
	wy = read(REG_WY)
	windowY = 0
}

func LcdBlitRow() {
	lcdc := read(REG_LCDC)
	y := ioLy()
	bufferBaseOffset := int(y) * LCD_WIDTH

	// render sprites if enabled
	if isBitSet(lcdc, LCDC_OBJ_ENABLE) {
		// get all visible sprites at this y
		sortedSprites := readSprites(y, isBitSet(lcdc, LCDC_OBJ_8X16))

		for _, sprite := range sortedSprites {
			for i, v := range sprite.pixels {
				offset := int(sprite.x) + i
				// off screen to the left
				if offset < SPRITE_X_OFFSET || offset >= LCD_WIDTH+SPRITE_X_OFFSET {
					continue
				}
				offset -= SPRITE_X_OFFSET

				currentValue := buffer[bufferBaseOffset+offset]
				if (v&0x0f) > 0 && (currentValue == 0 || isBitSet(currentValue, PIXEL_LOW_PRIORITY)) {
					buffer[bufferBaseOffset+offset] = v
				}
			}
		}
	}

	// render background if enabled
	if isBitSet(lcdc, LCDC_BG_WINDOW_ENABLE) {
		// get vram addresses
		bgMapBaseAddr := GPU_TILE_MAP_0
		if isBitSet(lcdc, LCDC_BG_TILE_MAP) {
			bgMapBaseAddr = GPU_TILE_MAP_1
		}
		windowMapBaseAddr := GPU_TILE_MAP_0
		if isBitSet(lcdc, LCDC_WINDOW_TILE_MAP) {
			windowMapBaseAddr = GPU_TILE_MAP_1
		}
		bgWindowDataBaseAddr := GPU_TILE_DATA_0
		if isBitSet(lcdc, LCDC_BG_WINDOW_TILE_DATA) {
			bgWindowDataBaseAddr = GPU_TILE_DATA_1
		}

		// get offsets
		scx := read(REG_SCX)
		scy := read(REG_SCY)
		bgY := uint8((int(y) + int(scy)) % LCD_BG_HEIGHT)

		wx := read(REG_WX)
		isShowWindow := isBitSet(lcdc, LCDC_WINDOW_ENABLE) && y >= wy && wy < LCD_HEIGHT && wx < LCD_WIDTH+WINDOW_X_OFFSET

		var x uint8
		for x = 0; x < LCD_WIDTH; x++ {
			// skip high priority sprites if already drawn
			currentValue := buffer[bufferBaseOffset+int(x)]
			if currentValue != 0 && !isBitSet(currentValue, PIXEL_LOW_PRIORITY) {
				continue
			}

			if isShowWindow && x+WINDOW_X_OFFSET >= wx {
				windowX := x + WINDOW_X_OFFSET - wx
				windowTileDataAddr := bgWindowTileDataAddr(windowMapBaseAddr, bgWindowDataBaseAddr, windowX, windowY)

				pixelX := uint8(windowX % 8)
				pixelY := uint8(windowY % 8)
				tilePaletteIndex := tilePaletteIndex(windowTileDataAddr, pixelX, pixelY)

				if tilePaletteIndex > 0 || currentValue == 0 {
					buffer[bufferBaseOffset+int(x)] = tilePaletteIndex | PIXEL_WINDOW
				}
			} else {
				bgX := uint8((int(x) + int(scx)) % LCD_BG_WIDTH)
				bgTileDataAddr := bgWindowTileDataAddr(bgMapBaseAddr, bgWindowDataBaseAddr, bgX, bgY)

				pixelX := uint8(bgX % 8)
				pixelY := uint8(bgY % 8)
				tilePaletteIndex := tilePaletteIndex(bgTileDataAddr, pixelX, pixelY)

				if tilePaletteIndex > 0 {
					buffer[bufferBaseOffset+int(x)] = tilePaletteIndex | PIXEL_BG
				}
			}
		}

		// window only increases y position if it is rendered
		if isShowWindow {
			windowY++
		}
	}

	// palettes
	bgWindowPalette := currentBgWindowPalette()
	obj0Palette := currentObjPalette(false)
	obj1Palette := currentObjPalette(true)

	colorMap := [16]uint8{
		bgWindowPalette[0], bgWindowPalette[1], bgWindowPalette[2], bgWindowPalette[3], // bg
		bgWindowPalette[0], bgWindowPalette[1], bgWindowPalette[2], bgWindowPalette[3], // window
		obj0Palette[0], obj0Palette[1], obj0Palette[2], obj0Palette[3], // obj0
		obj1Palette[0], obj1Palette[1], obj1Palette[2], obj1Palette[3], // obj1
	}

	// translate palette indexes to colours
	for x := 0; x < LCD_WIDTH; x++ {
		buffer[bufferBaseOffset+x] = colorMap[buffer[bufferBaseOffset+x]&0x0f]
	}
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

func bgWindowOffsetFromPixels(x, y uint8) uint16 {
	yOffset := uint16(y) / 8
	xOffset := uint16(x) / 8
	return yOffset*0x20 + xOffset
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
	sprites := make([]*SpriteData, 0, 5)

	var spriteHeight uint8 = 8
	if is8x16Mode {
		spriteHeight = 16
	}

	// filter sprites to ones on this line
	var i uint16
	for i = 0; i < 40; i++ {
		sprite := readSprite(i)

		if sprite.y == 0 || y+SPRITE_Y_OFFSET < sprite.y || y+SPRITE_Y_OFFSET-spriteHeight >= sprite.y {
			continue
		}

		if sprite.x >= SPRITE_X_OFFSET+LCD_WIDTH {
			continue
		}

		// read pixels
		spriteTileAddr := spriteTileAddr(sprite.tile)
		spriteY := y + SPRITE_Y_OFFSET - sprite.y
		if sprite.yFlip {
			spriteY = (spriteHeight - 1) - spriteY
		}

		spritePalette := PIXEL_OBJ0
		if sprite.palette {
			spritePalette = PIXEL_OBJ1
		}

		tileLo := read(spriteTileAddr + uint16(spriteY)*2)
		tileHi := read(spriteTileAddr + uint16(spriteY)*2 + 1)

		for j := 0; j < len(sprite.pixels); j++ {
			idx := j
			if sprite.xFlip {
				idx = (SPRITE_WIDTH - 1) - j
			}

			sprite.pixels[idx] = (tileLo >> uint(7-j)) & 0x01
			sprite.pixels[idx] |= ((tileHi >> uint(7-j)) & 0x01) << 1

			// 0 is transparent
			if sprite.pixels[idx] == 0 {
				continue
			}

			sprite.pixels[idx] |= spritePalette
			if sprite.priority {
				sprite.pixels[idx] |= PIXEL_LOW_PRIORITY
			}
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
		[8]uint8{},
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
