package ms3

import (
	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms1"
)

// Vec is a 3D vector. It is composed of 3 float32 fields for x, y, and z values in that order.
// Padding is added to reach 16 byte size for use with OpenGL with 1-to-1 mapping with vec3 and dvec3 types.
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Vec struct {
	X, Y, Z float32
	_       float32 // Padding for comfortable use with OpenGL. std430 still not fully supported on modern GPUs.
}

// Max returns the maximum component of a.
func (a Vec) Max() float32 {
	return math.Max(a.X, math.Max(a.Y, a.Z))
}

// Min returns the minimum component of a.
func (a Vec) Min() float32 {
	return math.Min(a.X, math.Min(a.Y, a.Z))
}

// Array returns the ordered components of Vec in a 3 element array [a.x,a.y,a.z].
func (a Vec) Array() [3]float32 {
	return [3]float32{a.X, a.Y, a.Z}
}

// AllNonzero returns true if all elements of a are nonzero.
func (a Vec) AllNonzero() bool {
	return a.X != 0 && a.Y != 0 && a.Z != 0
}

// Add returns the vector sum of p and q.
func Add(p, q Vec) Vec {
	return Vec{
		X: p.X + q.X,
		Y: p.Y + q.Y,
		Z: p.Z + q.Z,
	}
}

// AddScalar adds f to all of v's components and returns the result.
func AddScalar(f float32, v Vec) Vec {
	return Vec{
		X: v.X + f,
		Y: v.Y + f,
		Z: v.Z + f,
	}
}

// Sub returns the vector sum of p and -q.
func Sub(p, q Vec) Vec {
	return Vec{
		X: p.X - q.X,
		Y: p.Y - q.Y,
		Z: p.Z - q.Z,
	}
}

// Scale returns the vector p scaled by f.
func Scale(f float32, p Vec) Vec {
	return Vec{
		X: f * p.X,
		Y: f * p.Y,
		Z: f * p.Z,
	}
}

// Dot returns the dot product p·q.
func Dot(p, q Vec) float32 {
	return p.X*q.X + p.Y*q.Y + p.Z*q.Z
}

// Cross returns the cross product p×q.
func Cross(p, q Vec) Vec {
	return Vec{
		p.Y*q.Z - p.Z*q.Y,
		p.Z*q.X - p.X*q.Z,
		p.X*q.Y - p.Y*q.X, 0,
	}
}

// Norm returns the Euclidean norm of p
//
//	|p| = sqrt(p_x^2 + p_y^2 + p_z^2).
func Norm(p Vec) float32 {
	return math.Hypot(p.X, math.Hypot(p.Y, p.Z))
}

// Norm2 returns the Euclidean squared norm of p
//
//	|p|^2 = p_x^2 + p_y^2 + p_z^2.
func Norm2(p Vec) float32 {
	return p.X*p.X + p.Y*p.Y + p.Z*p.Z
}

// Unit returns the unit vector colinear to p.
// Unit returns {NaN,NaN,NaN} for the zero vector.
func Unit(p Vec) Vec {
	if p.X == 0 && p.Y == 0 && p.Z == 0 {
		return Vec{X: math.NaN(), Y: math.NaN(), Z: math.NaN()}
	}
	return Scale(1/Norm(p), p)
}

// Cos returns the cosine of the opening angle between p and q.
func Cos(p, q Vec) float32 {
	return Dot(p, q) / (Norm(p) * Norm(q))
}

// Divergence returns the divergence of the vector field at the point p,
// approximated using finite differences with the given step sizes.
func Divergence(p, step Vec, field func(Vec) Vec) float32 {
	sx := Vec{X: step.X}
	divx := (field(Add(p, sx)).X - field(Sub(p, sx)).X) / step.X
	sy := Vec{Y: step.Y}
	divy := (field(Add(p, sy)).Y - field(Sub(p, sy)).Y) / step.Y
	sz := Vec{Z: step.Z}
	divz := (field(Add(p, sz)).Z - field(Sub(p, sz)).Z) / step.Z
	return 0.5 * (divx + divy + divz)
}

// Gradient returns the gradient of the scalar field at the point p,
// approximated using finite differences with the given step sizes.
func Gradient(p, step Vec, field func(Vec) float32) Vec {
	dx := Vec{X: step.X}
	dy := Vec{Y: step.Y}
	dz := Vec{Z: step.Z}
	return Vec{
		X: (field(Add(p, dx)) - field(Sub(p, dx))) / (2 * step.X),
		Y: (field(Add(p, dy)) - field(Sub(p, dy))) / (2 * step.Y),
		Z: (field(Add(p, dz)) - field(Sub(p, dz))) / (2 * step.Z),
	}
}

// MinElem return a vector with the minimum components of two vectors.
func MinElem(a, b Vec) Vec {
	return Vec{
		X: math.Min(a.X, b.X),
		Y: math.Min(a.Y, b.Y),
		Z: math.Min(a.Z, b.Z),
	}
}

// MaxElem return a vector with the maximum components of two vectors.
func MaxElem(a, b Vec) Vec {
	return Vec{
		X: math.Max(a.X, b.X),
		Y: math.Max(a.Y, b.Y),
		Z: math.Max(a.Z, b.Z),
	}
}

// AbsElem returns the vector with components set to their absolute value.
func AbsElem(a Vec) Vec {
	return Vec{
		X: math.Abs(a.X),
		Y: math.Abs(a.Y),
		Z: math.Abs(a.Z),
	}
}

// MulElem returns the Hadamard product between vectors a and b.
//
//	v = {a.X*b.X, a.Y*b.Y, a.Z*b.Z}
func MulElem(a, b Vec) Vec {
	return Vec{
		X: a.X * b.X,
		Y: a.Y * b.Y,
		Z: a.Z * b.Z,
	}
}

// DivElem returns the Hadamard product between vector a
// and the inverse components of vector b.
//
//	v = {a.X/b.X, a.Y/b.Y, a.Z/b.Z}
func DivElem(a, b Vec) Vec {
	return Vec{
		X: a.X / b.X,
		Y: a.Y / b.Y,
		Z: a.Z / b.Z,
	}
}

// EqualElem checks equality between vector elements to within a tolerance.
func EqualElem(a, b Vec, tol float32) bool {
	return ms1.EqualWithinAbs(a.X, b.X, tol) &&
		ms1.EqualWithinAbs(a.Y, b.Y, tol) &&
		ms1.EqualWithinAbs(a.Z, b.Z, tol)
}

// elem returns a vector with all elements of magnitude length.
func elem(magnitude float32) Vec {
	return Vec{X: magnitude, Y: magnitude, Z: magnitude}
}

// Round rounds the individual elements of a vector.
func RoundElem(a Vec) Vec {
	return Vec{X: math.Round(a.X), Y: math.Round(a.Y), Z: math.Round(a.Z)}
}

// CeilElem returns a with Ceil applied to each component.
func CeilElem(a Vec) Vec {
	return Vec{X: math.Ceil(a.X), Y: math.Ceil(a.Y), Z: math.Ceil(a.Z)}
}

// FloorElem returns a with Floor applied to each component.
func FloorElem(a Vec) Vec {
	return Vec{X: math.Floor(a.X), Y: math.Floor(a.Y), Z: math.Floor(a.Z)}
}

// Sign returns sign function applied to each individual component of a. If a component is zero then zero is returned.
func SignElem(a Vec) Vec {
	return Vec{X: ms1.Sign(a.X), Y: ms1.Sign(a.Y), Z: ms1.Sign(a.Z)}
}

// SinElem returns sin(a) component-wise.
func SinElem(a Vec) Vec {
	return Vec{X: math.Sin(a.X), Y: math.Sin(a.Y), Z: math.Sin(a.Z)}
}

// CosElem returns cos(a) component-wise.
func CosElem(a Vec) Vec {
	return Vec{X: math.Cos(a.X), Y: math.Cos(a.Y), Z: math.Cos(a.Z)}
}

// SincosElem returns (sin(a), cos(a)). Is more efficient than calling both SinElem and CosElem.
func SincosElem(a Vec) (s, c Vec) {
	s.X, c.X = math.Sincos(a.X)
	s.Y, c.Y = math.Sincos(a.Y)
	s.Z, c.Z = math.Sincos(a.Z)
	return s, c
}

// Clamp returns v with its elements clamped to Min and Max's components.
func ClampElem(v, Min, Max Vec) Vec {
	return Vec{X: ms1.Clamp(v.X, Min.X, Max.X), Y: ms1.Clamp(v.Y, Min.Y, Max.Y), Z: ms1.Clamp(v.Z, Min.Z, Max.Z)}
}

// InterpElem performs a linear interpolation between x and y's elements, mapping with a's values in interval [0,1].
// This function is also known as "mix" in OpenGL.
func InterpElem(x, y, a Vec) Vec {
	return Vec{X: ms1.Interp(x.X, y.X, a.X), Y: ms1.Interp(x.Y, y.Y, a.Y), Z: ms1.Interp(x.Z, y.Z, a.Z)}
}

// SmoothStepElem performs element-wise smooth cubic hermite
// interpolation between 0 and 1 when e0 < x < e1.
func SmoothStepElem(e0, e1, x Vec) Vec {
	return Vec{X: ms1.SmoothStep(e0.X, e1.X, x.X), Y: ms1.SmoothStep(e0.Y, e1.Y, x.Y), Z: ms1.SmoothStep(e0.Z, e1.Z, x.Z)}
}
