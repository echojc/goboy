package main

import (
	"log"

	"github.com/deweerdt/gocui"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var texTileData1 uint32

func glCreateWindow() (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(320, 288, "goboy", nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	gl.ClearColor(1.0, 1.0, 1.0, 0.0)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0.0, 160.0, 144.0, 0.0, 1.0, -1.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &texTileData1)

	return window, nil
}

func glMainLoop(window *glfw.Window, g *gocui.Gui) {

	for !window.ShouldClose() && !guiCompleted {

		// rerender textures if necessary
		// only rerender if lcd is enabled
		if (read(0xff40) & 0x80) > 0 {
			if z80TileData0Dirty {
				UpdateTileData0()
			}
			if z80TileData1Dirty {
				UpdateTileData1()
			}
		}

		// clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// tile map 0, tile data 1
		gl.BindTexture(gl.TEXTURE_2D, texTileData1)
		var x, y int32
		for y = 0; y < 0x20; y++ {
			for x = 0; x < 0x20; x++ {
				addr := uint16(0x9800 + (y*0x20 + x))
				tile := read(addr)
				tileX := float32(tile&0x0f) / 16
				tileY := float32(tile>>4) / 16
				stride := float32(1) / 16

				gl.Begin(gl.QUADS)

				gl.TexCoord2f(tileX, tileY)
				gl.Vertex2i(x*8, y*8)

				gl.TexCoord2f(tileX, tileY+stride)
				gl.Vertex2i(x*8, (y+1)*8)

				gl.TexCoord2f(tileX+stride, tileY+stride)
				gl.Vertex2i((x+1)*8, (y+1)*8)

				gl.TexCoord2f(tileX+stride, tileY)
				gl.Vertex2i((x+1)*8, y*8)

				gl.End()
			}
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}

	// stop console too
	g.Execute(func(g *gocui.Gui) error {
		return gocui.ErrQuit
	})
}

func UpdateTileData0() {
	updateTileData(texTileData1, func(id uint8) uint16 {
		return 0x8000 + (uint16(id) << 4)
	})

	z80TileData0Dirty = false
	log.Println("Rendered tile data 0 to texture, resetting dirty flag")
}

func UpdateTileData1() {
	updateTileData(texTileData1, func(id uint8) uint16 {
		switch {
		case id < 0x80:
			return 0x9000 + (uint16(id) << 4)
		default:
			return 0x8000 + (uint16(id) << 4)
		}
	})

	z80TileData1Dirty = false
	log.Println("Rendered tile data 1 to texture, resetting dirty flag")
}

func updateTileData(textureId uint32, addrFunc func(uint8) uint16) {
	// cache palette as rgb values
	palette := getCurrentPalette()

	// decode texture data
	var texture [0x4000]uint8
	for id := 0; id < 0x100; id++ {
		addr := addrFunc(uint8(id))

		for y := 0; y < 8; y++ {
			// read bit values
			l := read(addr)
			addr++
			h := read(addr)
			addr++

			for x := 0; x < 8; x++ {
				bit := uint(7 - x)
				paletteIndex := ((l >> bit) & 0x01) | (((h >> bit) & 0x01) << 1)

				offsetY := (id>>4)*0x400 + y*0x80
				offsetX := (id&0x0f)*8 + x
				texture[offsetY+offsetX] = palette[paletteIndex]
			}
		}
	}

	// blit tile data
	gl.BindTexture(gl.TEXTURE_2D, textureId)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE, 128, 128, 0, gl.LUMINANCE, gl.UNSIGNED_BYTE, gl.Ptr(&texture[0]))
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
}

func getCurrentPalette() [4]uint8 {
	colors := [4]uint8{0xff, 0xaa, 0x55, 0x00}
	palette := read(0xff47)
	return [4]uint8{
		colors[palette&0x03],
		colors[(palette>>2)&0x03],
		colors[(palette>>4)&0x03],
		colors[(palette>>6)&0x03],
	}
}
