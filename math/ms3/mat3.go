package ms3

import (
	"errors"

	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms1"
)

// Mat3 is a 3x3 matrix.
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
type Mat3 struct {
	x00, x01, x02 float32
	x10, x11, x12 float32
	x20, x21, x22 float32
	// Padding to align to 16 bytes.
	_, _, _ float32
}

func mat3(x00, x01, x02, x10, x11, x12, x20, x21, x22 float32) Mat3 {
	return Mat3{
		x00, x01, x02,
		x10, x11, x12,
		x20, x21, x22, 0, 0, 0} // padding excess
}

// NewMat3 instantiates a new matrix from the first 9 floats, row major order. If v is of insufficient length NewMat3 panics.
func NewMat3(v []float32) (m Mat3) {
	_ = v[8]
	m.x00, m.x01, m.x02 = v[0], v[1], v[2]
	m.x10, m.x11, m.x12 = v[3], v[4], v[5]
	m.x20, m.x21, m.x22 = v[6], v[7], v[8]
	return m
}

// IdentityMat3 returns the 3x3 identity matrix.
func IdentityMat3() Mat3 {
	return mat3(
		1, 0, 0,
		0, 1, 0,
		0, 0, 1)
}

// Skew returns the 3×3 skew symmetric matrix (right hand system) of v.
//
//	                ⎡ 0 -z  y⎤
//	Skew({x,y,z}) = ⎢ z  0 -x⎥
//	                ⎣-y  x  0⎦
func Skew(v Vec) Mat3 {
	return mat3(
		0, -v.Z, v.Y,
		v.Z, 0, -v.X,
		-v.Y, v.X, 0)
}

// EqualMat3 tests the equality of 3x3 matrices.
func EqualMat3(a, b Mat3, tolerance float32) bool {
	return ms1.EqualWithinAbs(a.x00, b.x00, tolerance) &&
		ms1.EqualWithinAbs(a.x01, b.x01, tolerance) &&
		ms1.EqualWithinAbs(a.x02, b.x02, tolerance) &&
		ms1.EqualWithinAbs(a.x10, b.x10, tolerance) &&
		ms1.EqualWithinAbs(a.x11, b.x11, tolerance) &&
		ms1.EqualWithinAbs(a.x12, b.x12, tolerance) &&
		ms1.EqualWithinAbs(a.x20, b.x20, tolerance) &&
		ms1.EqualWithinAbs(a.x21, b.x21, tolerance) &&
		ms1.EqualWithinAbs(a.x22, b.x22, tolerance)
}

// MulPosition multiplies a V2 position with a rotate/translate matrix.
func (a Mat3) mulPosition(x, y float32) (float32, float32) {
	return a.x00*x + a.x01*y + a.x02,
		a.x10*x + a.x11*y + a.x12
}

// MulMat3 multiplies two 3x3 matrices.
func MulMat3(a, b Mat3) Mat3 {
	m := Mat3{}
	m.x00 = a.x00*b.x00 + a.x01*b.x10 + a.x02*b.x20
	m.x10 = a.x10*b.x00 + a.x11*b.x10 + a.x12*b.x20
	m.x20 = a.x20*b.x00 + a.x21*b.x10 + a.x22*b.x20
	m.x01 = a.x00*b.x01 + a.x01*b.x11 + a.x02*b.x21
	m.x11 = a.x10*b.x01 + a.x11*b.x11 + a.x12*b.x21
	m.x21 = a.x20*b.x01 + a.x21*b.x11 + a.x22*b.x21
	m.x02 = a.x00*b.x02 + a.x01*b.x12 + a.x02*b.x22
	m.x12 = a.x10*b.x02 + a.x11*b.x12 + a.x12*b.x22
	m.x22 = a.x20*b.x02 + a.x21*b.x12 + a.x22*b.x22
	return m
}

// AddMat3 adds two 3x3 matrices together and returns the result.
func AddMat3(a, b Mat3) Mat3 {
	return Mat3{
		x00: a.x00 + b.x00,
		x10: a.x10 + b.x10,
		x20: a.x20 + b.x20,
		x01: a.x01 + b.x01,
		x11: a.x11 + b.x11,
		x21: a.x21 + b.x21,
		x02: a.x02 + b.x02,
		x12: a.x12 + b.x12,
		x22: a.x22 + b.x22,
	}
}

// SubMat3 subtracts a 3x3 matrix b from a andf returns the result.
func SubMat3(a, b Mat3) Mat3 {
	return Mat3{
		x00: a.x00 - b.x00,
		x10: a.x10 - b.x10,
		x20: a.x20 - b.x20,
		x01: a.x01 - b.x01,
		x11: a.x11 - b.x11,
		x21: a.x21 - b.x21,
		x02: a.x02 - b.x02,
		x12: a.x12 - b.x12,
		x22: a.x22 - b.x22,
	}
}

// Prod performs vector multiplication as if they were matrices
//
//	m = v1 * v2ᵀ
func Prod(v1, v2t Vec) Mat3 {
	return mat3(
		v1.X*v2t.X, v1.X*v2t.Y, v1.X*v2t.Z,
		v1.Y*v2t.X, v1.Y*v2t.Y, v1.Y*v2t.Z,
		v1.Z*v2t.X, v1.Z*v2t.Y, v1.Z*v2t.Z,
	)
}

// MulMatVec performs matrix multiplication on v:
//
//	result = M * v
func MulMatVec(m Mat3, v Vec) (result Vec) {
	result.X = v.X*m.x00 + v.Y*m.x01 + v.Z*m.x02
	result.Y = v.X*m.x10 + v.Y*m.x11 + v.Z*m.x12
	result.Z = v.X*m.x20 + v.Y*m.x21 + v.Z*m.x22
	return result
}

// MulMatVecTrans Performs transposed matrix multiplication on v:
//
//	result = Mᵀ * v
func MulMatVecTrans(m Mat3, v Vec) (result Vec) {
	result.X = v.X*m.x00 + v.Y*m.x10 + v.Z*m.x20
	result.Y = v.X*m.x01 + v.Y*m.x11 + v.Z*m.x21
	result.Z = v.X*m.x02 + v.Y*m.x12 + v.Z*m.x22
	return result
}

// ScaleMat3 multiplies each 3x3 matrix component by a scalar.
func ScaleMat3(a Mat3, k float32) Mat3 {
	return Mat3{
		x00: k * a.x00,
		x10: k * a.x10,
		x20: k * a.x20,
		x01: k * a.x01,
		x11: k * a.x11,
		x21: k * a.x21,
		x02: k * a.x02,
		x12: k * a.x12,
		x22: k * a.x22,
	}
}

// Determinant returns the determinant of a 3x3 matrix.
func (a Mat3) Determinant() float32 {
	return (a.x00*(a.x11*a.x22-a.x21*a.x12) -
		a.x01*(a.x10*a.x22-a.x20*a.x12) +
		a.x02*(a.x10*a.x21-a.x20*a.x11))
}

// Inverse returns the inverse of a 3x3 matrix.
func (a Mat3) Inverse() Mat3 {
	m := Mat3{}
	det := a.Determinant()
	if det == 0 {
		return mat3(math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN())
	}
	d := 1.0 / det
	m.x00 = (a.x11*a.x22 - a.x12*a.x21) * d
	m.x01 = (a.x21*a.x02 - a.x01*a.x22) * d
	m.x02 = (a.x01*a.x12 - a.x11*a.x02) * d
	m.x10 = (a.x12*a.x20 - a.x22*a.x10) * d
	m.x11 = (a.x22*a.x00 - a.x20*a.x02) * d
	m.x12 = (a.x02*a.x10 - a.x12*a.x00) * d
	m.x20 = (a.x10*a.x21 - a.x20*a.x11) * d
	m.x21 = (a.x20*a.x01 - a.x00*a.x21) * d
	m.x22 = (a.x00*a.x11 - a.x01*a.x10) * d
	return m
}

// Transpose returns the transpose of a.
func (a Mat3) Transpose() Mat3 {
	return Mat3{
		x00: a.x00, x01: a.x10, x02: a.x20,
		x10: a.x01, x11: a.x11, x12: a.x21,
		x20: a.x02, x21: a.x12, x22: a.x22,
	}
}

// VecDiag returns the matrix diagonal as a Vec.
func (m Mat3) VecDiag() Vec {
	return Vec{X: m.x00, Y: m.x11, Z: m.x22}
}

// VecRow returns the ith row as a Vec.
func (m Mat3) VecRow(i int) Vec {
	switch i {
	case 0:
		return Vec{X: m.x00, Y: m.x01, Z: m.x02}
	case 1:
		return Vec{X: m.x10, Y: m.x11, Z: m.x12}
	case 2:
		return Vec{X: m.x20, Y: m.x21, Z: m.x22}
	}
	panic("out of bounds")
}

// VecCol returns the jth column as a Vec.
func (m Mat3) VecCol(j int) Vec {
	switch j {
	case 0:
		return Vec{X: m.x00, Y: m.x10, Z: m.x20}
	case 1:
		return Vec{X: m.x01, Y: m.x11, Z: m.x21}
	case 2:
		return Vec{X: m.x02, Y: m.x12, Z: m.x22}
	}
	panic("out of bounds")
}

// Put stores the matrix values into slice b in row major order. If b is not of length 9 or greater Put panics.
func (m Mat3) Put(b []float32) {
	_ = b[8]
	b[0], b[1], b[2] = m.x00, m.x01, m.x02
	b[3], b[4], b[5] = m.x10, m.x11, m.x12
	b[6], b[7], b[8] = m.x20, m.x21, m.x22
}

// Array returns the matrix values in a static array copy in row major order.
func (m Mat3) Array() (rowmajor [9]float32) {
	m.Put(rowmajor[:])
	return rowmajor
}

// AsMat4 expands the Mat3 to fill the first rows and columns of a Mat4
// and sets the last diagonal element of the Mat4 to 1.
func (m Mat3) AsMat4() Mat4 {
	return Mat4{
		m.x00, m.x01, m.x02, 0,
		m.x10, m.x11, m.x12, 0,
		m.x20, m.x21, m.x22, 0,
		0, 0, 0, 1,
	}
}

// RotatingMat3 returns a 3×3 rotation matrix corresponding to the receiver. It
// may be used to perform rotations on a 3-vector or to apply the rotation
// to a 3×n matrix of column vectors. If the receiver is not a unit
// quaternion, the returned matrix will not be a pure rotation.
func RotatingMat3(rotationUnit Quat) Mat3 {
	w, i, j, k := rotationUnit.W, rotationUnit.I, rotationUnit.J, rotationUnit.K
	ii := 2 * i * i
	jj := 2 * j * j
	kk := 2 * k * k
	wi := 2 * w * i
	wj := 2 * w * j
	wk := 2 * w * k
	ij := 2 * i * j
	jk := 2 * j * k
	ki := 2 * k * i
	return mat3(
		1-(jj+kk), ij-wk, ki+wj,
		ij+wk, 1-(ii+kk), jk-wi,
		ki-wj, jk+wi, 1-(ii+jj))
}

// Hessian returns the Hessian matrix of the vector field f at point p.
// step is the step with which the second derivative is calculated.
func Hessian(p Vec, step float32, f func(Vec) float32) Mat3 {
	h2 := step * step * 4
	dx := Vec{X: step}
	dy := Vec{Y: step}
	dz := Vec{Z: step}
	fp := f(p)
	diff2 := func(p, d1, d2 Vec, f func(p Vec) float32) float32 {
		return (f(Add(p, Add(d1, d2))) - f(Add(p, d2)) - f(Add(p, d1)) + fp) / h2
	}
	fxx := diff2(p, dx, dx, f)
	fyy := diff2(p, dy, dy, f)
	fzz := diff2(p, dz, dz, f)
	fxy := diff2(p, dx, dy, f)
	fxz := diff2(p, dx, dz, f)
	fyz := diff2(p, dy, dz, f)
	return mat3(
		fxx, fxy, fxz,
		fxy, fyy, fyz,
		fxz, fyz, fzz,
	)
}

// Eigs returns the real and imaginary parts of the 3 eigenvalues of m. It returns a non-nil error if it is unable to solve.
func (m Mat3) Eigs() (r, c [3]float32, err error) {
	const tol = 1e-12
	if !ms1.EqualWithinAbs(m.x01, m.x10, tol) ||
		!ms1.EqualWithinAbs(m.x12, m.x21, tol) ||
		!ms1.EqualWithinAbs(m.x02, m.x20, tol) {
		return r, c, errors.New("non-symmetric eigenvalue algorithm not implemented")
	}
	// 3*m = tr(A)
	M := (m.x00 + m.x11 + m.x22) / 3
	// Calculate  2*q=det(A-m*I)
	nm := ScaleMat3(IdentityMat3(), M)
	nm = SubMat3(m, nm)
	q := nm.Determinant() / 2
	// 6*p = sum of squares of elements of A-m*I
	const sixdiv = 1. / 6
	var p float32
	for _, v := range nm.Array() {
		p += sixdiv * v * v
	}

	if math.Abs(p) < tol && math.Abs(q) < tol {
		// p == q == 0
		return [3]float32{M, M, M}, [3]float32{}, nil
	}
	// sqrt(3)
	const sqrt3 = 1.7320508075688772935274463415058723669428052538103806280558069794
	// phi = 1/3 atan( sqrt(p^3 - q^2)/q ), 0<=phi<=pi
	phi := math.Atan(math.Sqrt(p*p*p-q*q)/q) / 3
	sp, cp := math.Sincos(phi)
	sqrtp := math.Sqrt(p)
	return [3]float32{
		M + 2*sqrtp*cp,
		M - sqrtp*(cp+sqrt3*sp),
		M - sqrtp*(cp-sqrt3*sp),
	}, [3]float32{}, nil
}
