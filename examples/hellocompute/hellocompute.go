// This program adds a uniform value to
// an array using a compute shader.
package main

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"

	"log/slog"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/soypat/glgl/v4.6-core/glgl"
)

type def struct {
	//go:embed hellocompute.glsl
	data string
}

const (
	width   = 20
	height  = 20
	addThis = 99
)

func init() {
	fmt.Println(def{})
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

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
		slog.Error("initializing", "err", err.Error())
		return
	}
	defer terminate()
	fmt.Println(glgl.MaxComputeWorkGroupCount())
	fmt.Println(glgl.MaxComputeWorkGroupSize())
	return
	ss, err := glgl.ParseCombined(strings.NewReader(compute))
	if err != nil {
		slog.Error("parsing", "err", err.Error())
		return
	}
	prog, err := glgl.CompileProgram(ss)
	if err != nil {
		slog.Error("creating program", "err", err.Error())
		return
	}
	prog.Bind()
	err = prog.SetUniform1f("u_adder\x00", addThis)
	if err != nil {
		slog.Error("setting uniform", "err", err.Error())
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
		ImageUnit:      unit,
	}
	// DST starts with ones, and we add the uniform variable too all values.
	dst := make([]float32, width*height)
	for i := range dst {
		dst[i]++
	}
	tex, err := glgl.NewTextureFromImage(cfg, dst)
	if err != nil {
		slog.Error("creating texture", "err", err.Error())
		return
	}

	// Dispatch and wait for compute to finish.
	err = prog.RunCompute(width, height, 1)
	if err != nil {
		slog.Error("running compute shader", "err", err.Error())
		return
	}

	err = glgl.GetImage(dst, tex, cfg)
	if err != nil {
		slog.Error("acquiring results from GPU", "err", err.Error())
		return
	}
	fmt.Println(dst)
}
