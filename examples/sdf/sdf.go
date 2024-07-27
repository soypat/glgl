// This program takes input array and a add scalar
// and writes the result of adding them to a new output array.
// The value of this example is that input and output arrays
// may have different memory layouts. One may contain vector data
// and the other may contain scalars.
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"math"
	"os"
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
const fltFmtByte = 'g'

type SDFShader struct {
	Name []byte
	Body []byte
}

func (s *Sphere) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error { return nil }

func (s *Sphere) AppendShader(glsl *SDFShader) error {
	r := float64(s.R)
	glsl.Name = append(glsl.Name, "sphere"...)
	glsl.Name = strconv.AppendFloat(glsl.Name, r, fltFmtByte, fltPrec, 32)
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
		r := norm(pos)
		distances[i] = r - s.R
	}
	return 0, nil
}

func (s *Sphere) Bounds() (min, max Vec) {
	min = Vec{X: s.R, Y: s.R, Z: s.R}
	max = Vec{X: s.R, Y: s.R, Z: s.R}
	return min, max
}

type SDFShaderer interface {
	Bounds() (min, max Vec)
	AppendShader(glsl *SDFShader) error
	ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error
}

type BinaryOpShader struct {
	s1, s2  SDFShaderer
	opname  string
	bodyFmt string
}

func Union(s1, s2 SDFShaderer) SDFShaderer {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &UnionShader{
		s1: s1,
		s2: s2,
	}
}

type UnionShader struct {
	s1, s2 SDFShaderer
}

func (s *UnionShader) Bounds() (vmin, vmax Vec) {
	min1, max1 := s.s1.Bounds()
	min2, max2 := s.s2.Bounds()
	vmin = Vec{X: minf(min1.X, min2.X), Y: minf(min1.Y, min2.Y), Z: minf(min1.Z, min2.Z)}
	vmax = Vec{X: maxf(max1.X, max2.X), Y: maxf(max1.Y, max2.Y), Z: maxf(max1.Z, max2.Z)}
	return vmin, vmax
}

func (s *UnionShader) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error {
	err := fn(flags, s.s1)
	if err != nil {
		return err
	}
	return fn(flags, s.s2)
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

func Translate(s SDFShaderer, to Vec) SDFShaderer {
	return &TranslateShader{
		s: s,
		p: to,
	}
}

type TranslateShader struct {
	s SDFShaderer
	p Vec
}

func (ts *TranslateShader) Bounds() (min, max Vec) {
	min, max = ts.s.Bounds()
	min = Vec{X: min.X + ts.p.X, Y: min.Y + ts.p.Y, Z: min.X + ts.p.Z}
	max = Vec{X: max.X + ts.p.X, Y: max.Y + ts.p.Y, Z: max.X + ts.p.Z}
	return min, max
}

func (s *TranslateShader) ForEachChild(flags int, fn func(flags int, s SDFShaderer) error) error {
	return fn(flags, s.s)
}

func (ts *TranslateShader) AppendShader(glsl *SDFShader) error {
	glsl.Name = append(glsl.Name, "translate"...)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.X), fltFmtByte, fltPrec, 32)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.Y), fltFmtByte, fltPrec, 32)
	glsl.Name = strconv.AppendFloat(glsl.Name, float64(ts.p.Z), fltFmtByte, fltPrec, 32)
	for {
		idx := bytes.IndexByte(glsl.Name, '.')
		if idx < 0 {
			break
		}
		glsl.Name[idx] = 'p'
	}
	glsl.Name = append(glsl.Name, '_')
	idStart := len(glsl.Name)
	body := glsl.Body
	err := ts.s.AppendShader(glsl)
	if err != nil {
		return err
	}
	glsl.Body = glsl.Body[:len(body)]
	glsl.Body = append(glsl.Body, "return "...)
	glsl.Body = append(glsl.Body, glsl.Name[idStart:]...)
	glsl.Body = append(glsl.Body, "(p - vec3("...)
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.X), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, ',')
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.Y), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, ',')
	glsl.Body = strconv.AppendFloat(glsl.Body, float64(ts.p.Z), 'f', fltPrec, 32)
	glsl.Body = append(glsl.Body, "));"...)
	return nil
}

func main() {
	s1, _ := NewSphere(0.5)
	s2, _ := NewSphere(1)
	s1 = Translate(s1, Vec{X: 2})
	obj := Union(s1, s2)

	fp, err := os.Create("sdf_gen.glsl")
	if err != nil {
		panic(err)
	}
	writeProgram(fp, obj)
}

func writeProgram(w io.Writer, obj SDFShaderer) (n int, err error) {
	Children := []SDFShaderer{obj}
	nextChild := 0
	for len(Children[nextChild:]) > 0 {
		prev := len(Children)
		for _, obj := range Children[nextChild:] {
			fmt.Printf("%T\n", obj)
			obj.ForEachChild(0, func(flags int, s SDFShaderer) error {
				Children = append(Children, s)
				return nil
			})
		}
		nextChild = prev
	}
	n, err = w.Write([]byte("#shader compute\n#version 430\n\n"))
	if err != nil {
		return n, err
	}
	var scratch SDFShader
	for i := len(Children) - 1; i >= 0; i-- {
		ngot, err := writeShader(w, Children[i], &scratch)
		n += ngot
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func writeShader(w io.Writer, s SDFShaderer, scratch *SDFShader) (int, error) {
	scratch.Name = scratch.Name[:0]
	scratch.Body = scratch.Body[:0]
	scratch.Name = append(scratch.Name, "float "...)
	err := s.AppendShader(scratch)
	if err != nil {
		return 0, err
	}
	scratch.Name = append(scratch.Name, "(vec3 p) {\n"...)

	scratch.Body = append(scratch.Body, "\n}\n\n"...)
	n, err := w.Write(scratch.Name)
	if err != nil {
		return n, err
	}
	n2, err := w.Write(scratch.Body)
	return n + n2, err
}

func minf(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func maxf(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

// norm is equivalent to glsl `length` call.
func norm(pos Vec) float32 {
	r1 := math.Hypot(float64(pos.X), float64(pos.Y))
	r2 := math.Hypot(r1, float64(pos.Z))
	return float32(r2)
}
