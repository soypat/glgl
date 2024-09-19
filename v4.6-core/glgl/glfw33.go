//go:build !tinygo && cgo

package glgl

import (
	"errors"

	"log/slog"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type WindowConfig struct {
	Title        string
	NotResizable bool
	Version      [2]int
	// glfw.OpenGLCoreProfile
	OpenGLProfile int
	ForwardCompat bool
	Width, Height int
	DebugLog      *slog.Logger
}

func InitWithCurrentWindow33(cfg WindowConfig) (*glfw.Window, func(), error) {
	if cfg.DebugLog != nil {
		return nil, nil, errors.New("DebugLog not supported in GLFW version 3.3")
	}
	if err := glfw.Init(); err != nil {
		return nil, nil, err
	}

	glfw.WindowHint(glfw.Resizable, b2i(!cfg.NotResizable))
	if cfg.Version != [2]int{} {
		glfw.WindowHint(glfw.ContextVersionMajor, cfg.Version[0])
		glfw.WindowHint(glfw.ContextVersionMinor, cfg.Version[1])
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 6)
	}
	glfw.WindowHint(glfw.OpenGLProfile, zdefault(cfg.OpenGLProfile, glfw.OpenGLCoreProfile))
	glfw.WindowHint(glfw.OpenGLForwardCompatible, b2i(cfg.ForwardCompat))
	window, err := glfw.CreateWindow(cfg.Width, cfg.Height, cfg.Title, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		glfw.Terminate()
		return window, nil, err
	}
	ClearErrors()
	return window, glfw.Terminate, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
