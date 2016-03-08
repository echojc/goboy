package main

import (
	"github.com/deweerdt/gocui"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

func glCreateWindow() (*glfw.Window, error) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(160, 144, "goboy", nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	gl.Enable(gl.DEPTH_TEST)

	gl.ClearColor(1.0, 1.0, 1.0, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0.0, 160.0, 144.0, 0.0, 1.0, -1.0)

	return window, nil
}

func glMainLoop(window *glfw.Window, g *gocui.Gui) {
	var r float32

	for !window.ShouldClose() && !guiCompleted {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		gl.Translatef(80, 72, 0)
		gl.Rotatef(r, 0, 0, 1)
		gl.Translatef(-80, -72, 0)
		r += 1

		gl.Begin(gl.TRIANGLES)
		gl.Color3f(1, 0, 0)
		gl.Vertex2f(80, 20)
		gl.Color3f(0, 1, 0)
		gl.Vertex2f(50, 120)
		gl.Color3f(0, 0, 1)
		gl.Vertex2f(110, 120)
		gl.End()

		window.SwapBuffers()
		glfw.PollEvents()
	}

	// stop console too
	g.Execute(func(g *gocui.Gui) error {
		return gocui.ErrQuit
	})
}
