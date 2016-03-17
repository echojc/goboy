package main

import (
	"fmt"

	"github.com/deweerdt/gocui"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const windowTitleFormat = "goboy vfps:%.01f cfps:%.01f"

var glFps Fps
var glFrameCount uint64
var glCurrentScreenBuffer []uint8
var glScreenDirty bool

var texScreen uint32

func glCreateWindow() (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	title := fmt.Sprintf(windowTitleFormat, 0.0, 0.0)
	window, err := glfw.CreateWindow(LCD_WIDTH*2, LCD_HEIGHT*2, title, nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	gl.ClearColor(1.0, 1.0, 1.0, 0.0)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0.0, LCD_WIDTH, LCD_HEIGHT, 0.0, 1.0, -1.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Enable(gl.TEXTURE_2D)
	gl.GenTextures(1, &texScreen)

	return window, nil
}

func glMainLoop(window *glfw.Window, g *gocui.Gui) {

	for !window.ShouldClose() && !guiCompleted {
		lcdc := read(REG_LCDC)

		// clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// only render if lcd is enabled
		if isBitSet(lcdc, LCDC_ENABLE) {

			if glScreenDirty {
				glUpdateScreen()
			}
			gl.BindTexture(gl.TEXTURE_2D, texScreen)

			// render bg
			gl.Begin(gl.QUADS)

			gl.TexCoord2f(0, 0)
			gl.Vertex2i(0, 0)

			gl.TexCoord2f(0, 1)
			gl.Vertex2i(0, LCD_HEIGHT)

			gl.TexCoord2f(1, 1)
			gl.Vertex2i(LCD_WIDTH, LCD_HEIGHT)

			gl.TexCoord2f(1, 0)
			gl.Vertex2i(LCD_WIDTH, 0)

			gl.End()
		}

		glFrameCount++
		if glFrameCount%60 == 0 {
			glFps.Add(glFrameCount)

			title := fmt.Sprintf(windowTitleFormat, glFps.Current(), z80Fps.Current())
			window.SetTitle(title)
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}

	// stop console too
	g.Execute(func(g *gocui.Gui) error {
		return gocui.ErrQuit
	})
}

func glSetBuffer(buffer []uint8) {
	glCurrentScreenBuffer = buffer
	glScreenDirty = true
}

func glUpdateScreen() {
	buffer := glCurrentScreenBuffer

	gl.BindTexture(gl.TEXTURE_2D, texScreen)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE, LCD_WIDTH, LCD_HEIGHT, 0, gl.LUMINANCE, gl.UNSIGNED_BYTE, gl.Ptr(&buffer[0]))
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	glScreenDirty = false
}
