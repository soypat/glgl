package ms3

import (
	math "github.com/chewxy/math32"
)

// Triangle represents a triangle in 3D space and
// is composed by 3 vectors corresponding to the position
// of each of the vertices. Ordering of these vertices
// decides the "normal" direction.
// Inverting ordering of two vertices inverts the resulting direction.
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Triangle [3]Vec

// Centroid returns the intersection of the three medians of the triangle
// as a point in space.
func (t Triangle) Centroid() Vec {
	return Scale(1.0/3.0, Add(Add(t[0], t[1]), t[2]))
}

// Normal returns the vector with direction
// perpendicular to the Triangle's face and magnitude
// twice that of the Triangle's area. The ordering
// of the triangle vertices decides the normal's resulting
// direction. The returned vector is not normalized.
func (t Triangle) Normal() Vec {
	s1, s2, _ := t.sides()
	return Cross(s1, s2)
}

// IsDegenerate returns true if all of triangle's vertices are
// within tol distance of its longest side.
func (t Triangle) IsDegenerate(tol float32) bool {
	longIdx := t.longIdx()
	// calculate vertex distance from longest side
	ln := Line{t[longIdx], t[(longIdx+1)%3]}
	dist := ln.Distance(t[(longIdx+2)%3])
	return dist <= tol
}

// longIdx returns index of the longest side. The sides
// of the triangles are are as follows:
//   - Side 0 formed by vertices 0 and 1
//   - Side 1 formed by vertices 1 and 2
//   - Side 2 formed by vertices 0 and 2
func (t Triangle) longIdx() int {
	sides := [3]Vec{Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])}
	len2 := [3]float32{Norm2(sides[0]), Norm2(sides[1]), Norm2(sides[2])}
	longLen := len2[0]
	longIdx := 0
	if len2[1] > longLen {
		longLen = len2[1]
		longIdx = 1
	}
	if len2[2] > longLen {
		longIdx = 2
	}
	return longIdx
}

// Area returns the surface area of the triangle.
func (t Triangle) Area() float32 {
	// Heron's Formula, see https://en.wikipedia.org/wiki/Heron%27s_formula.
	// Also see William M. Kahan (24 March 2000). "Miscalculating Area and Angles of a Needle-like Triangle"
	// for more discussion. https://people.eecs.berkeley.edu/~wkahan/Triangle.pdf.
	a, b, c := t.orderedLengths()
	A := (c + (b + a)) * (a - (c - b))
	A *= (a + (c - b)) * (c + (b - a))
	return math.Sqrt(A) / 4
}

// Sort performs the sort-3 algorithm and returns
// l1, l2, l3 such that l1 ≤ l2 ≤ l3.
func Sort(a, b, c float32) (l1, l2, l3 float32) {
	// sort-3
	if l2 < l1 {
		l1, l2 = l2, l1
	}
	if l3 < l2 {
		l2, l3 = l3, l2
		if l2 < l1 {
			l1, l2 = l2, l1
		}
	}
	return l1, l2, l3
}

// orderedLengths returns the lengths of the sides of the triangle such that
// a ≤ b ≤ c.
func (t Triangle) orderedLengths() (a, b, c float32) {
	s1, s2, s3 := t.sides()
	l1 := Norm(s1)
	l2 := Norm(s2)
	l3 := Norm(s3)
	return Sort(l1, l2, l3)
}

// sides returns vectors for each of the sides of t.
func (t Triangle) sides() (Vec, Vec, Vec) {
	return Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])
}

// Line is an infinite 3D line
// defined by two points on the line.
type Line [2]Vec

// Interpolate takes a value between 0 and 1 to linearly
// interpolate a point on the line.
//
//	Interpolate(0) returns l[0]
//	Interpolate(1) returns l[1]
func (l Line) Interpolate(t float32) Vec {
	lineDir := Sub(l[1], l[0])
	return Add(l[0], Scale(t, lineDir))
}

// Distance returns the minimum euclidean Distance of point p
// to the infinite line.
func (l Line) Distance(p Vec) float32 {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	num := Norm(Cross(Sub(p, l[0]), Sub(p, l[1])))
	return num / Norm(Sub(l[1], l[0]))
}
