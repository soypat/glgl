// This program takes input array and a add scalar
// and writes the result of adding them to a new output array.
// The value of this example is that input and output arrays
// may have different memory layouts. One may contain vector data
// and the other may contain scalars.
package main

import (
	"bytes"
	_ "embed"
	"math"
	"runtime"
	"strconv"
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

const fltPrec = 8

type SDFShader struct {
	Name []byte
	Body []byte
}

func (s *Sphere) AppendShader(glsl *SDFShader) error {
	r := float64(s.R)
	glsl.Name = append(glsl.Name, "sphere_"...)
	glsl.Name = strconv.AppendFloat(glsl.Name, r, 'f', fltPrec, 32)
	if idx := bytes.IndexByte(glsl.Name, '.'); idx >= 0 {
		// Identifiers cannot have period in name.
		glsl.Name[idx] = 'p'
	}
	glsl.Body = append(glsl.Body, "return length(p)-"...)
	glsl.Body = strconv.AppendFloat(glsl.Body, r, 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, ';')
	return nil
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

type SDFShaderer interface {
	AppendShader(glsl *SDFShader) error
}

type BinaryOpShader struct {
	s1, s2  SDFShaderer
	opname  string
	bodyFmt string
}

type UnionShader struct {
	s1, s2 SDFShaderer
}

func (s *UnionShader) AppendShader(glsl *SDFShader) error {
	body := glsl.Body
	glsl.Name = append(glsl.Name, "union_"...)
	id1Start := len(glsl.Name)
	err := s.s1.AppendShader(glsl)
	if err != nil {
		return err
	}
	id2Start := len(glsl.Name)
	err = s.s2.AppendShader(glsl)
	if err != nil {
		return err
	}
	glsl.Body = glsl.Body[:len(body)] // Remove union element bodies but retain longer
	glsl.Body = append(glsl.Body, "return min("...)
	glsl.Body = append(glsl.Body, glsl.Name[id1Start:id2Start]...)
	glsl.Body = append(glsl.Body, "(p),"...)
	glsl.Body = append(glsl.Body, glsl.Name[id2Start:]...)
	glsl.Body = append(glsl.Body, "(p));"...)
	return nil
}

func NewSphere(radius float32) (SDFShaderer, error) {
	return &Sphere{R: radius}, nil
}

func main() {

}
