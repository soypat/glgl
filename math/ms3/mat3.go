package ms3

import (
	math "github.com/chewxy/math32"
)

// Mat3 is a 3x3 matrix.
type Mat3 struct {
	x00, x01, x02 float32
	x10, x11, x12 float32
	x20, x21, x22 float32
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
	return Mat3{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1}
}

// Skew returns the 3×3 skew symmetric matrix (right hand system) of v.
//
//	                ⎡ 0 -z  y⎤
//	Skew({x,y,z}) = ⎢ z  0 -x⎥
//	                ⎣-y  x  0⎦
func Skew(v Vec) Mat3 {
	return Mat3{
		0, -v.Z, v.Y,
		v.Z, 0, -v.X,
		-v.Y, v.X, 0}
}

// EqualMat3 tests the equality of 3x3 matrices.
func EqualMat3(a, b Mat3, tolerance float32) bool {
	return (math.Abs(a.x00-b.x00) < tolerance &&
		math.Abs(a.x01-b.x01) < tolerance &&
		math.Abs(a.x02-b.x02) < tolerance &&
		math.Abs(a.x10-b.x10) < tolerance &&
		math.Abs(a.x11-b.x11) < tolerance &&
		math.Abs(a.x12-b.x12) < tolerance &&
		math.Abs(a.x20-b.x20) < tolerance &&
		math.Abs(a.x21-b.x21) < tolerance &&
		math.Abs(a.x22-b.x22) < tolerance)
}

// MulPosition multiplies a V2 position with a rotate/translate matrix.
func (a Mat3) mulPosition(x, y float32) (float32, float32) {
	return a.x00*x + a.x01*y + a.x02,
		a.x10*x + a.x11*y + a.x12
}

// MulMat3 multiplies 3x3 matrices.
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

// AddMat3 two 3x3 matrices.
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

// ProdMat3 performs vector multiplication as if they were matrices
//
//	m = v1 * v2ᵀ
func Prod(v1, v2t Vec) Mat3 {
	return Mat3{
		v1.X * v2t.X, v1.X * v2t.Y, v1.X * v2t.Z,
		v1.Y * v2t.X, v1.Y * v2t.Y, v1.Y * v2t.Z,
		v1.Z * v2t.X, v1.Z * v2t.Y, v1.Z * v2t.Z,
	}
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
		return Mat3{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
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

func (m Mat3) expand() Mat4 {
	return Mat4{
		m.x00, m.x01, m.x02, 0,
		m.x10, m.x11, m.x12, 0,
		m.x20, m.x21, m.x22, 0,
		0, 0, 0, 1,
	}
}

// Mat returns a 3×3 rotation matrix corresponding to the receiver. It
// may be used to perform rotations on a 3-vector or to apply the rotation
// to a 3×n matrix of column vectors. If the receiver is not a unit
// quaternion, the returned matrix will not be a pure rotation.
func RotationMat3(rotationUnit Quat) Mat3 {
	w, i, j, k := rotationUnit.W, rotationUnit.V.X, rotationUnit.V.Y, rotationUnit.V.Z
	ii := 2 * i * i
	jj := 2 * j * j
	kk := 2 * k * k
	wi := 2 * w * i
	wj := 2 * w * j
	wk := 2 * w * k
	ij := 2 * i * j
	jk := 2 * j * k
	ki := 2 * k * i
	return Mat3{
		1 - (jj + kk), ij - wk, ki + wj,
		ij + wk, 1 - (ii + kk), jk - wi,
		ki - wj, jk + wi, 1 - (ii + jj)}
}
