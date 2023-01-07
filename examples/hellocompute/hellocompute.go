package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"golang.org/x/exp/slog"
)

const (
	width   = 20
	height  = 20
	addThis = 99
)

//go:embed hellocompute.glsl
var compute string

func main() {
	_, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   width,
		Height:  height,
	})
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	ss, err := glgl.ParseCombined(strings.NewReader(compute))
	if err != nil {
		panic(err)
	}
	prog, err := glgl.NewProgram(ss)
	if err != nil {
		slog.Error("creating program", err)
		return
	}
	prog.Bind()
	const unit = 0
	cfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          width,
		Height:         height,
		Access:         glgl.ReadOrWrite,
		Format:         gl.RED,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.R32F,
		Unit:           unit,
	}
	dst := make([]float32, width*height)
	tex, err := glgl.NewTextureFromImage(cfg, dst)
	if err != nil {
		panic(err)
	}
	// Dispatch and wait for compute to finish.
	err = prog.RunCompute(width, height, 1)
	if err != nil {
		panic(err)
	}

	err = glgl.GetImage(dst, tex, cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(dst)
}
