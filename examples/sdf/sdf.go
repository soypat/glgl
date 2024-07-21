// This program takes input array and a add scalar
// and writes the result of adding them to a new output array.
// The value of this example is that input and output arrays
// may have different memory layouts. One may contain vector data
// and the other may contain scalars.
package main

import (
	_ "embed"
	"math"
	"runtime"
	"strconv"
	"strings"
)

const addThis = 20

var (
	//go:embed sdf.glsl
	compute string
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

type Vec struct {
	X, Y, Z float32
}

type SDF interface {
	Evaluate(positions []Vec, distances []float32) (int, error)
	Bounds() (min, max Vec)
}

type Sphere struct {
	R float32
}

type SDFShader struct {
	Name string
	Body string
}

func (s *Sphere) Shader() SDFShader {
	r := strconv.FormatFloat(float64(s.R), 'f', 16, 32)
	if !strings.Contains(r, ".") {
		r += "."
	}
	return SDFShader{
		Name: "sphere_" + r,
		Body: "float r=" + r + ";return length(p)-r;",
	}
}

func (s *Sphere) Evaluate(positions []Vec, distances []float32) (int, error) {
	for i, pos := range positions {
		r1 := math.Hypot(float64(pos.X), float64(pos.Y))
		r2 := math.Hypot(r1, float64(pos.Z))
		distances[i] = float32(r2) - s.R
	}
	return 0, nil
}

func (s *Sphere) Bounds() (min, max Vec) {
	min = Vec{X: s.R, Y: s.R, Z: s.R}
	max = Vec{X: s.R, Y: s.R, Z: s.R}
	return min, max
}

func NewSphere(radius float32) (SDF, error) {
	return &Sphere{R: radius}, nil
}

func main() {

}
