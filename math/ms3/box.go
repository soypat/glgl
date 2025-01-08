package ms3

import (
	math "github.com/chewxy/math32"
)

// Box is a 3D bounding box. Well formed Boxes Min components
// are smaller than Max components. Max is the most positive/largest vertex,
// Min is the most negative/smallest vertex.
type Box struct {
	Min, Max Vec
}

// NewBox is shorthand for Box{Min:Vec{x0,y0,z0}, Max:Vec{x1,y1,z1}}.
// The sides are swapped so that the resulting Box is well formed.
func NewBox(x0, y0, z0, x1, y1, z1 float32) Box {
	return Box{
		Min: Vec{X: math.Min(x0, x1), Y: math.Min(y0, y1), Z: math.Min(z0, z1)},
		Max: Vec{X: math.Max(x0, x1), Y: math.Max(y0, y1), Z: math.Max(z0, z1)},
	}
}

// NewCenteredBox returns a box centered around center with size dimensions.
func NewCenteredBox(center, size Vec) Box {
	size = MaxElem(size, Vec{}) // set negative values to zero.
	half := Scale(0.5, size)
	return Box{Min: Sub(center, half), Max: Add(center, half)}
}

// IsEmpty returns true if a Box's volume is zero
// or if a Min component is greater than its Max component.
func (a Box) Empty() bool {
	return a.Min.X >= a.Max.X || a.Min.Y >= a.Max.Y || a.Min.Z >= a.Max.Z
}

// Size returns the size of the Box.
func (a Box) Size() Vec {
	return Sub(a.Max, a.Min)
}

// Center returns the center of the Box.
func (a Box) Center() Vec {
	return Scale(0.5, Add(a.Min, a.Max))
}

// Volume returns the volume contained within the box. Returns 0 for malformed boxes.
func (a Box) Volume() float32 {
	sz := a.Size()
	if sz.X < 0 || sz.Y < 0 || sz.Z < 0 {
		return 0
	}
	return sz.X * sz.Z * sz.Y
}

// Vertices returns a slice of the 8 vertices
// corresponding to each of the Box's corners.
//
// Ordering of vertices 0-3 is CCW in the XY plane starting at box minimum.
// Ordering of vertices 4-7 is CCW in the XY plane starting at box minimum
// for X and Y values and maximum Z value.
//
// Edges for the box can be constructed with the following indices:
//
//	edges := [12][2]int{
//	 {0, 1}, {1, 2}, {2, 3}, {3, 0},
//	 {4, 5}, {5, 6}, {6, 7}, {7, 4},
//	 {0, 4}, {1, 5}, {2, 6}, {3, 7},
//	}
func (a Box) Vertices() [8]Vec {
	return [8]Vec{
		0: a.Min,
		1: {X: a.Max.X, Y: a.Min.Y, Z: a.Min.Z},
		2: {X: a.Max.X, Y: a.Max.Y, Z: a.Min.Z},
		3: {X: a.Min.X, Y: a.Max.Y, Z: a.Min.Z},
		4: {X: a.Min.X, Y: a.Min.Y, Z: a.Max.Z},
		5: {X: a.Max.X, Y: a.Min.Y, Z: a.Max.Z},
		6: a.Max,
		7: {X: a.Min.X, Y: a.Max.Y, Z: a.Max.Z},
	}
}

// Union returns a box enclosing both the receiver and argument Boxes.
func (a Box) Union(b Box) Box {
	if a.Empty() {
		return b
	} else if b.Empty() {
		return a
	}
	return Box{
		Min: MinElem(a.Min, b.Min),
		Max: MaxElem(a.Max, b.Max),
	}
}

// Intersect returns a box enclosing the box space shared by both boxes.
func (a Box) Intersect(b Box) (intersect Box) {
	// Calculate the intersection minimum and maximum coordinates using MinElem and MaxElem
	intersect.Min = MaxElem(a.Min, b.Min)
	intersect.Max = MinElem(a.Max, b.Max)
	if intersect.Empty() {
		return Box{}
	}
	return intersect
}

// IncludePoint returns a box containing both the receiver and the argument point.
func (a Box) IncludePoint(point Vec) Box {
	return Box{
		Min: MinElem(a.Min, point),
		Max: MaxElem(a.Max, point),
	}
}

// Add adds v to the bounding box components.
// It is the equivalent of translating the Box by v.
func (a Box) Add(v Vec) Box {
	return Box{Min: Add(a.Min, v), Max: Add(a.Max, v)}
}

// ScaleCentered returns a new Box scaled by a size vector around its center.
// The scaling is done element wise which is to say the Box's X dimension
// is scaled by scale.X. Negative elements of scale are interpreted as zero.
func (a Box) ScaleCentered(scale Vec) Box {
	scale = MaxElem(scale, Vec{})
	// TODO(soypat): Probably a better way to do this.
	return NewCenteredBox(a.Center(), MulElem(scale, a.Size()))
}

// Scale scales the box dimensions and positions in 3 directions. Does not preserve box center.
// Negative elements of scale can be used to mirror box dimension.
func (a Box) Scale(scale Vec) Box {
	newCenter := MulElem(scale, a.Center())
	return NewCenteredBox(newCenter, MulElem(AbsElem(scale), a.Size()))
}

// Contains returns true if v is contained within the bounds of the Box.
func (a Box) Contains(point Vec) bool {
	if a.Empty() {
		return point == a.Min && point == a.Max
	}
	return a.Min.X <= point.X && point.X <= a.Max.X &&
		a.Min.Y <= point.Y && point.Y <= a.Max.Y &&
		a.Min.Z <= point.Z && point.Z <= a.Max.Z
}

// ContainsBox returns true if argument box is fully contained within receiver box.
func (a Box) ContainsBox(b Box) bool { return a.Contains(b.Min) && a.Contains(b.Max) }

// Equal returns true if a and b are within tol of eachother for each box limit component.
func (a Box) Equal(b Box, tol float32) bool {
	return EqualElem(a.Min, b.Min, tol) && EqualElem(a.Max, b.Max, tol)
}

// Canon returns the canonical version of a. The returned Box has minimum
// and maximum coordinates swapped if necessary so that it is well-formed.
func (a Box) Canon() Box {
	return Box{
		Min: MinElem(a.Min, a.Max),
		Max: MaxElem(a.Min, a.Max),
	}
}

// Diagonal returns a's diagonal length: sqrt(L*L + W*W + H*H).
func (a Box) Diagonal() float32 {
	sz := a.Size()
	return math.Hypot(math.Hypot(sz.X, sz.Y), sz.Z)
}
