package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func ExportTileMap0() {
	file, err := os.Create("/tmp/tile_map_0.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	out := image.NewGray(image.Rect(0, 0, 32*8, 32*8))

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			tileIndex := read(uint16(0x9800 + (y*32 + x)))
			copyTile(out, getTile(tileIndex), x, y)
		}
	}

	png.Encode(file, out)
}

func ExportTileData() {
	file, err := os.Create("/tmp/tile_data.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	out := image.NewGray(image.Rect(0, 0, 16*8, 24*8))

	for y := 0; y < 24; y++ {
		for x := 0; x < 16; x++ {
			copyTile(out, convertTile(uint16(0x8000+((y*16+x)<<4))), x, y)
		}
	}

	png.Encode(file, out)
}

func copyTile(out *image.Gray, tile *image.Gray, x int, y int) {
	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			out.SetGray((x*8)+i, (y*8)+j, tile.GrayAt(i, j))
		}
	}
}

func getTile(index uint8) *image.Gray {
	// TODO read which tile memory bank to use
	var addr uint16
	switch {
	case index < 0x80:
		addr = 0x9000 + (uint16(index) << 4)
	default:
		addr = 0x8000 + (uint16(index) << 4)
	}
	return convertTile(addr)
}

func convertTile(addr uint16) *image.Gray {
	out := image.NewGray(image.Rect(0, 0, 8, 8))

	var y uint16
	for y = 0; y < 8; y++ {
		line := convertTileLine(read(addr+(2*y)), read(addr+(2*y+1)))
		for x, v := range line {
			out.SetGray(x, int(y), v)
		}
	}

	return out
}

func convertTileLine(l uint8, h uint8) [8]color.Gray {
	var out [8]color.Gray

	for i := 0; i < 8; i++ {
		bit := uint(7 - i)
		paletteIndexL := ((l >> bit) & 0x01)
		paletteIndexH := ((h >> bit) & 0x01) << 1
		out[i] = convertTilePalette(paletteIndexH | paletteIndexL)
	}

	return out
}

func convertTilePalette(i uint8) color.Gray {
	palette := read(0xff47)
	colorMapIndex := (palette >> (i * 2)) & 0x03
	return colorMap[colorMapIndex]
}

var colorMap = [4]color.Gray{
	color.Gray{0xff},
	color.Gray{0xaa},
	color.Gray{0x55},
	color.Gray{0x00},
}
