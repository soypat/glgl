// This program takes input array and a add scalar
// and writes the result of adding them to a new output array.
// The value of this example is that input and output arrays
// may have different memory layouts. One may contain vector data
// and the other may contain scalars.
package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"golang.org/x/exp/slog"
)

const addThis = 20

var (
	// Contains the input data.
	inputArray = []float32{1, 2, 3, 4, 5}
	// Will contain the output data.
	outputArray = make([]float32, len(inputArray))
	//go:embed pipecompute.glsl
	compute string
)

func main() {
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   1,
		Height:  1,
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
	// Unit must match the `binding` of the texture in the compute shader.
	const (
		inputUnit  = 0
		outputUnit = 2
	)
	inputCfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          len(inputArray),
		Height:         1,
		Access:         glgl.ReadOnly,
		Format:         gl.RED,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.R32F,
		TextureUnit:    inputUnit,
	}
	_, err = glgl.NewTextureFromImage(inputCfg, inputArray)
	if err != nil {
		slog.Error("creating input texture", err)
		return
	}

	// Define OUTPUT texture.
	outputCfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          len(outputArray),
		Height:         1,
		Access:         glgl.ReadOnly,
		Format:         gl.RED,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.R32F,
		TextureUnit:    outputUnit,
	}
	outputTex, err := glgl.NewTextureFromImage(outputCfg, outputArray)
	if err != nil {
		slog.Error("creating output texture", err)
		return
	}

	// Dispatch and wait for compute to finish.
	err = prog.RunCompute(len(inputArray), 1, 1)
	if err != nil {
		slog.Error("running compute shader", err)
		return
	}

	err = glgl.GetImage(outputArray, outputTex, outputCfg)
	if err != nil {
		slog.Error("acquiring results from GPU", err)
		return
	}
	fmt.Println("input:", inputArray, "output:", outputArray)
}
