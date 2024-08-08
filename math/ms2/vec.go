package ms2

import (
	math "github.com/chewxy/math32"
)

// Vec is a 3D vector. It is composed of 3 float32 fields for x, y, and z values in that order.
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
	return Vec{X: sign(a.X), Y: sign(a.Y)}
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
	return Vec{X: clamp(v.X, Min.X, Max.X), Y: clamp(v.Y, Min.Y, Max.Y)}
}

// InterpElem performs a linear interpolation between x and y's elements, mapping with a's values in interval [0,1].
// This function is also known as "mix" in OpenGL.
func InterpElem(x, y, a Vec) Vec {
	return Vec{X: interp(x.X, y.X, a.X), Y: clamp(x.Y, y.Y, a.Y)}
}

// sign returns -1, 0, or 1 for negative, zero or positive x argument, respectively, just like OpenGL's "sign" function.
func sign(x float32) float32 {
	if x == 0 {
		return 0
	}
	return math.Copysign(1, x)
}

// clamp returns value v clamped between Min and Max.
func clamp(v, Min, Max float32) float32 {
	return math.Min(Max, math.Max(v, Min))
}

// interp performs the linear interpolation between x and y, mapping with a in interval [0,1].
// This function is known as "mix" in OpenGL.
func interp(x, y, a float32) float32 {
	return x*(1-a) + y*a
}
