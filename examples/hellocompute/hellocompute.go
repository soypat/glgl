// This program adds a uniform value to
// an array using a compute shader.
package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
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
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   width,
		Height:  height,
	})
	if err != nil {
		slog.Error("initializing", err)
		return
	}
	defer terminate()

	ss, err := glgl.ParseCombined(strings.NewReader(compute))
	if err != nil {
		slog.Error("parsing", err)
		return
	}
	prog, err := glgl.NewProgram(ss)
	if err != nil {
		slog.Error("creating program", err)
		return
	}
	prog.Bind()
	err = prog.SetUniform1f("u_adder\x00", addThis)
	if err != nil {
		slog.Error("setting uniform", err)
		return
	}
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
		slog.Error("creating texture", err)
		return
	}
	// Dispatch and wait for compute to finish.
	err = prog.RunCompute(width, height, 1)
	if err != nil {
		slog.Error("running compute shader", err)
		return
	}

	err = glgl.GetImage(dst, tex, cfg)
	if err != nil {
		slog.Error("acquiring results from GPU", err)
		return
	}
	fmt.Println(dst)
}
