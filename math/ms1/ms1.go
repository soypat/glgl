// package ms1 implements basic 1D math useful for 3D graphics applications.
// Functions in this package have their OpenGL equivalent which is usually of the same name.
package ms1

import math "github.com/chewxy/math32"

// Sign returns -1, 0, or 1 for negative, zero or positive x argument, respectively, just like OpenGL's "sign" function.
func Sign(x float32) float32 {
	if x == 0 {
		return 0
	}
	return math.Copysign(1, x)
}

// Clamp returns value v clamped between Min and Max.
func Clamp(v, Min, Max float32) float32 {
	return math.Min(Max, math.Max(v, Min))
}

// Interp performs the linear interpolation between x and y, mapping with a in interval [0,1].
// This function is known as "mix" in OpenGL.
func Interp(x, y, a float32) float32 {
	return x*(1-a) + y*a
}

// SmoothStep performs smooth cubic hermite interpolation between 0 and 1 when edge0 < x < edge1.
func SmoothStep(edge0, edge1, x float32) float32 {
	t := Clamp((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * (3 - 2*t)
}
