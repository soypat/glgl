package ms2

import (
	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms1"
)

// Vec is a 2D vector. It is composed of 2 float32 fields for x and y values in that order.
type Vec struct {
	X, Y float32
}

// Max returns the maximum component of a.
func (a Vec) Max() float32 {
	return math.Max(a.X, a.Y)
}

// Min returns the minimum component of a.
func (a Vec) Min() float32 {
	return math.Min(a.X, a.Y)
}

// Array returns the ordered components of Vec in a 2 element array [a.x,a.y].
func (a Vec) Array() [2]float32 {
	return [2]float32{a.X, a.Y}
}

// AllNonzero returns true if all elements of a are nonzero.
func (a Vec) AllNonzero() bool {
	return a.X != 0 && a.Y != 0
}

// Add returns the vector sum of p and q.
func Add(p, q Vec) Vec {
	return Vec{
		X: p.X + q.X,
		Y: p.Y + q.Y,
	}
}

// AddScalar adds f to all of v's components and returns the result.
func AddScalar(f float32, v Vec) Vec {
	return Vec{
		X: v.X + f,
		Y: v.Y + f,
	}
}

// Sub returns the vector sum of p and -q.
func Sub(p, q Vec) Vec {
	return Vec{
		X: p.X - q.X,
		Y: p.Y - q.Y,
	}
}

// Scale returns the vector p scaled by f.
func Scale(f float32, p Vec) Vec {
	return Vec{
		X: f * p.X,
		Y: f * p.Y,
	}
}

// Cross returns the cross product p×q.
func Cross(p, q Vec) float32 {
	return p.X*q.Y - p.Y*q.X
}

// Dot returns the dot product p·q.
func Dot(p, q Vec) float32 {
	return p.X*q.X + p.Y*q.Y
}

// Norm returns the Euclidean norm of p
//
//	|p| = sqrt(p_x^2 + p_y^2).
func Norm(p Vec) float32 {
	return math.Hypot(p.X, p.Y)
}

// Norm2 returns the Euclidean squared norm of p
//
//	|p|^2 = p_x^2 + p_y^2
func Norm2(p Vec) float32 {
	return p.X*p.X + p.Y*p.Y
}

// Unit returns the unit vector colinear to p.
// Unit returns {NaN,NaN,NaN} for the zero vector.
func Unit(p Vec) Vec {
	if p.X == 0 && p.Y == 0 {
		return Vec{X: math.NaN(), Y: math.NaN()}
	}
	return Scale(1/Norm(p), p)
}

// Cos returns the cosine of the opening angle between p and q.
func Cos(p, q Vec) float32 {
	return Dot(p, q) / (Norm(p) * Norm(q))
}

// MinElem return a vector with the minimum components of two vectors.
func MinElem(a, b Vec) Vec {
	return Vec{
		X: math.Min(a.X, b.X),
		Y: math.Min(a.Y, b.Y),
	}
}

// MaxElem return a vector with the maximum components of two vectors.
func MaxElem(a, b Vec) Vec {
	return Vec{
		X: math.Max(a.X, b.X),
		Y: math.Max(a.Y, b.Y),
	}
}

// AbsElem returns the vector with components set to their absolute value.
func AbsElem(a Vec) Vec {
	return Vec{
		X: math.Abs(a.X),
		Y: math.Abs(a.Y),
	}
}

// MulElem returns the Hadamard product between vectors a and b.
//
//	v = {a.X*b.X, a.Y*b.Y}
func MulElem(a, b Vec) Vec {
	return Vec{
		X: a.X * b.X,
		Y: a.Y * b.Y,
	}
}

// DivElem returns the Hadamard product between vector a
// and the inverse components of vector b.
//
//	v = {a.X/b.X, a.Y/b.Y}
func DivElem(a, b Vec) Vec {
	return Vec{
		X: a.X / b.X,
		Y: a.Y / b.Y,
	}
}

// EqualElem checks equality between vector elements to within a tolerance.
func EqualElem(a, b Vec, tol float32) bool {
	return math.Abs(a.X-b.X) <= tol &&
		math.Abs(a.Y-b.Y) <= tol
}

// elem returns a vector with all elements of magnitude length.
func elem(magnitude float32) Vec {
	return Vec{X: magnitude, Y: magnitude}
}

// Round rounds the individual elements of a vector.
func RoundElem(a Vec) Vec {
	return Vec{X: math.Round(a.X), Y: math.Round(a.Y)}
}

// CeilElem returns a with Ceil applied to each component.
func CeilElem(a Vec) Vec {
	return Vec{X: math.Ceil(a.X), Y: math.Ceil(a.Y)}
}

// FloorElem returns a with Floor applied to each component.
func FloorElem(a Vec) Vec {
	return Vec{X: math.Floor(a.X), Y: math.Floor(a.Y)}
}

// Sign returns sign function applied to each individual component of a. If a component is zero then zero is returned.
func SignElem(a Vec) Vec {
	return Vec{X: ms1.Sign(a.X), Y: ms1.Sign(a.Y)}
}

// SinElem returns sin(a) component-wise.
func SinElem(a Vec) Vec {
	return Vec{X: math.Sin(a.X), Y: math.Sin(a.Y)}
}

// CosElem returns cos(a) component-wise.
func CosElem(a Vec) Vec {
	return Vec{X: math.Cos(a.X), Y: math.Cos(a.Y)}
}

// SincosElem returns (sin(a), cos(a)). Is more efficient than calling both SinElem and CosElem.
func SincosElem(a Vec) (s, c Vec) {
	s.X, c.X = math.Sincos(a.X)
	s.Y, c.Y = math.Sincos(a.Y)
	return s, c
}

// Clamp returns v with its elements clamped to Min and Max's components.
func ClampElem(v, Min, Max Vec) Vec {
	return Vec{X: ms1.Clamp(v.X, Min.X, Max.X), Y: ms1.Clamp(v.Y, Min.Y, Max.Y)}
}

// InterpElem performs a linear interpolation between x and y's elements, mapping with a's values in interval [0,1].
// This function is also known as "mix" in OpenGL.
func InterpElem(x, y, a Vec) Vec {
	return Vec{X: ms1.Interp(x.X, y.X, a.X), Y: ms1.Interp(x.Y, y.Y, a.Y)}
}

// pol is a polar coordinate tuple.
type pol struct {
	R     float32
	Theta float32
}

// Cartesian converts polar coordinates p to cartesian coordinates.
func (p pol) Cartesian() Vec {
	return Vec{X: p.R * math.Cos(p.Theta), Y: p.R * math.Sin(p.Theta)}
}

// polar converts cartesian coordinates v to polar coordinates.
func (v Vec) polar() pol {
	return pol{Norm(v), math.Atan2(v.Y, v.X)}
}

// SmoothStepElem performs element-wise smooth cubic hermite
// interpolation between 0 and 1 when e0 < x < e1.
func SmoothStepElem(e0, e1, x Vec) Vec {
	return Vec{X: ms1.SmoothStep(e0.X, e1.X, x.X), Y: ms1.SmoothStep(e0.Y, e1.Y, x.Y)}
}

// Orientation calculates the orientation in the plane of 3 points and applies it to f.
//   - f returned for counter-clockwise orientation
//   - -f returned for clockwise orientation
//   - 0 returned for 3 colinear points
func CopyOrientation(f float32, p1, p2, p3 Vec) float32 {
	// See C++ version: https://www.geeksforgeeks.org/orientation-3-ordered-points/
	slope1 := (p2.Y - p1.Y) * (p3.X - p2.X)
	slope2 := (p3.Y - p2.Y) * (p2.X - p1.X)
	if slope1 == slope2 {
		return 0
	}
	return math.Copysign(f, slope2-slope1)
}

// Collinear returns true if 3 points lie on a single line to within tol.
func Collinear(a, b, c Vec, tol float32) bool {
	pa := Unit(Sub(a, c))
	pb := Unit(Sub(b, c))
	return math.Abs(Cross(pa, pb)) < tol
}
