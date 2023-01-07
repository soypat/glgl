package glgl

import (
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type WindowConfig struct {
	Title         string
	NotResizable  bool
	Version       [2]int
	OpenGLProfile int
	ForwardCompat bool
	Width, Height int
}

func InitWithCurrentWindow33(cfg WindowConfig) (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.Resizable, b2i(!cfg.NotResizable))
	if cfg.Version != [2]int{} {
		glfw.WindowHint(glfw.ContextVersionMajor, cfg.Version[0])
		glfw.WindowHint(glfw.ContextVersionMinor, cfg.Version[1])
	}
	if cfg.OpenGLProfile != 0 {
		glfw.WindowHint(glfw.OpenGLProfile, cfg.OpenGLProfile)
	}

	glfw.WindowHint(glfw.OpenGLForwardCompatible, b2i(cfg.ForwardCompat))
	window, err := glfw.CreateWindow(cfg.Width, cfg.Height, cfg.Title, nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		glfw.Terminate()
		return window, err
	}
	ClearErrors()
	return window, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
