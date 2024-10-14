package ms3_test

import (
	"math/rand"
	"testing"
	"unsafe"

	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
)

func BenchmarkSVD(b *testing.B) {
	a := ms3.NewMat3([]float32{
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
	})
	for i := 0; i < b.N; i++ {
		_, _, _ = a.SVD()
	}
}

func BenchmarkSVDOriginal(b *testing.B) {
	a := ms3.NewMat3([]float32{
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
		float32(rand.Float64()), float32(rand.Float64()), float32(rand.Float64()),
	})
	for i := 0; i < b.N; i++ {
		_, _, _ = SVD(a)
	}
}

func TestSVD(t *testing.T) {
	const tol = 1e-5
	rng := rand.New(rand.NewSource(1))
	random := func() float32 {
		return float32(rng.Float64() * 10)
	}
	for i := 0; i < 100000; i++ {
		a := ms3.NewMat3([]float32{
			random(), random(), random(),
			random(), random(), random(),
			random(), random(), random(),
		})
		if i%8 == 0 {
			a = a.Inverse()
		} else if i%8 == 1 {
			a = ms3.MulMat3(a.Transpose(), a)
		}
		uwant, swant, vwant := SVD(a)
		ugot, sgot, vgot := a.SVD()
		if !ms3.EqualMat3(uwant, ugot, tol) {
			t.Error("U mismatch")
		}
		if !ms3.EqualMat3(swant, sgot, tol) {
			t.Error("S mismatch")
		}
		if !ms3.EqualMat3(vwant, vgot, tol) {
			t.Error("V mismatch")
		}
	}
}

// Constants used in the algorithm
const (
	gamma   = 5.828427124  // FOUR_GAMMA_SQUARED = sqrt(8)+3
	cstar   = 0.923879532  // cos(pi/8)
	sstar   = 0.3826834323 // sin(pi/8)
	epsilon = 1e-6
)

// rsqrt computes an approximate reciprocal square root of x.
func rsqrt(x float32) float32 {
	xhalf := 0.5 * x
	i := math.Float32bits(x)
	i = 0x5f375a82 - (i >> 1)
	x = math.Float32frombits(i)
	x = x * (1.5 - xhalf*x*x)
	return x
}

// rsqrt1 computes a more accurate reciprocal square root of x.
func rsqrt1(x float32) float32 {
	xhalf := 0.5 * x
	i := math.Float32bits(x)
	i = 0x5f37599e - (i >> 1)
	x = math.Float32frombits(i)
	x = x * (1.5 - xhalf*x*x)
	x = x * (1.5 - xhalf*x*x)
	return x
}

// accurateSqrt computes the square root of x using rsqrt1.
func accurateSqrt(x float32) float32 {
	return x * rsqrt1(x)
}

// condSwap swaps X and Y if condition c is true.
func condSwap(c bool, X, Y *float32) {
	if c {
		*X, *Y = *Y, *X
	}
}

// condNegSwap swaps X and Y and negates X if condition c is true.
func condNegSwap(c bool, X, Y *float32) {
	if c {
		*X, *Y = *Y, -*X
	}
}

// multAB computes the matrix multiplication M = A * B.
func multAB(
	a11, a12, a13,
	a21, a22, a23,
	a31, a32, a33 float32,
	b11, b12, b13,
	b21, b22, b23,
	b31, b32, b33 float32,
	m11, m12, m13,
	m21, m22, m23,
	m31, m32, m33 *float32,
) {
	*m11 = a11*b11 + a12*b21 + a13*b31
	*m12 = a11*b12 + a12*b22 + a13*b32
	*m13 = a11*b13 + a12*b23 + a13*b33
	*m21 = a21*b11 + a22*b21 + a23*b31
	*m22 = a21*b12 + a22*b22 + a23*b32
	*m23 = a21*b13 + a22*b23 + a23*b33
	*m31 = a31*b11 + a32*b21 + a33*b31
	*m32 = a31*b12 + a32*b22 + a33*b32
	*m33 = a31*b13 + a32*b23 + a33*b33
}

// multAtB computes the matrix multiplication M = Transpose[A] * B.
func multAtB(
	a11, a12, a13,
	a21, a22, a23,
	a31, a32, a33 float32,
	b11, b12, b13,
	b21, b22, b23,
	b31, b32, b33 float32,
	m11, m12, m13,
	m21, m22, m23,
	m31, m32, m33 *float32,
) {
	*m11 = a11*b11 + a21*b21 + a31*b31
	*m12 = a11*b12 + a21*b22 + a31*b32
	*m13 = a11*b13 + a21*b23 + a31*b33
	*m21 = a12*b11 + a22*b21 + a32*b31
	*m22 = a12*b12 + a22*b22 + a32*b32
	*m23 = a12*b13 + a22*b23 + a32*b33
	*m31 = a13*b11 + a23*b21 + a33*b31
	*m32 = a13*b12 + a23*b22 + a33*b32
	*m33 = a13*b13 + a23*b23 + a33*b33
}

// quatToMat3 converts a quaternion to a 3x3 rotation matrix.
func quatToMat3(qV [4]float32,
	m11, m12, m13,
	m21, m22, m23,
	m31, m32, m33 *float32,
) {
	w := qV[3]
	x := qV[0]
	y := qV[1]
	z := qV[2]

	qxx := x * x
	qyy := y * y
	qzz := z * z
	qxz := x * z
	qxy := x * y
	qyz := y * z
	qwx := w * x
	qwy := w * y
	qwz := w * z

	*m11 = 1 - 2*(qyy+qzz)
	*m12 = 2 * (qxy - qwz)
	*m13 = 2 * (qxz + qwy)
	*m21 = 2 * (qxy + qwz)
	*m22 = 1 - 2*(qxx+qzz)
	*m23 = 2 * (qyz - qwx)
	*m31 = 2 * (qxz - qwy)
	*m32 = 2 * (qyz + qwx)
	*m33 = 1 - 2*(qxx+qyy)
}

// approximateGivensQuaternion computes the Givens rotation quaternion.
func approximateGivensQuaternion(a11, a12, a22 float32, ch, sh *float32) {
	*ch = 2 * (a11 - a22)
	*sh = a12
	b := gamma*(*sh)*(*sh) < (*ch)*(*ch)
	w := rsqrt((*ch)*(*ch) + (*sh)*(*sh))
	if b {
		*ch = w * (*ch)
		*sh = w * (*sh)
	} else {
		*ch = cstar
		*sh = sstar
	}
}

// jacobiConjugation performs the Jacobi rotation to diagonalize the matrix.
func jacobiConjugation(x, y, z int,
	s11 *float32,
	s21, s22 *float32,
	s31, s32, s33 *float32,
	qV *[4]float32) {

	var ch, sh float32
	approximateGivensQuaternion(*s11, *s21, *s22, &ch, &sh)

	scale := ch*ch + sh*sh
	a := (ch*ch - sh*sh) / scale
	b := (2 * sh * ch) / scale

	// Make temp copy of S
	_s11 := *s11
	_s21 := *s21
	_s22 := *s22
	_s31 := *s31
	_s32 := *s32
	_s33 := *s33

	// Perform conjugation S = Q'*S*Q
	*s11 = a*(a*_s11+b*_s21) + b*(a*_s21+b*_s22)
	*s21 = a*(-b*_s11+a*_s21) + b*(-b*_s21+a*_s22)
	*s22 = -b*(-b*_s11+a*_s21) + a*(-b*_s21+a*_s22)
	*s31 = a*_s31 + b*_s32
	*s32 = -b*_s31 + a*_s32
	*s33 = _s33

	// Update cumulative rotation qV
	var tmp [3]float32
	tmp[0] = qV[0] * sh
	tmp[1] = qV[1] * sh
	tmp[2] = qV[2] * sh
	sh *= qV[3]

	qV[0] *= ch
	qV[1] *= ch
	qV[2] *= ch
	qV[3] *= ch

	// (x,y,z) corresponds to ((0,1,2),(1,2,0),(2,0,1))
	qV[z] += sh
	qV[3] -= tmp[z]
	qV[x] += tmp[y]
	qV[y] -= tmp[x]

	// Rearrange matrix for next iteration
	_s11 = *s22
	_s21 = *s32
	_s22 = *s33
	_s31 = *s21
	_s32 = *s31
	_s33 = *s11
	*s11 = _s11
	*s21 = _s21
	*s22 = _s22
	*s31 = _s31
	*s32 = _s32
	*s33 = _s33
}

// dist2 computes the squared distance.
func dist2(x, y, z float32) float32 {
	return x*x + y*y + z*z
}

// jacobiEigenanalysis diagonalizes a symmetric matrix using Jacobi rotations.
func jacobiEigenanalysis(
	s11 *float32,
	s21, s22 *float32,
	s31, s32, s33 *float32,
	qV *[4]float32,
) {
	qV[3] = 1
	qV[0] = 0
	qV[1] = 0
	qV[2] = 0
	for i := 0; i < 4; i++ {
		jacobiConjugation(0, 1, 2, s11, s21, s22, s31, s32, s33, qV)
		jacobiConjugation(1, 2, 0, s11, s21, s22, s31, s32, s33, qV)
		jacobiConjugation(2, 0, 1, s11, s21, s22, s31, s32, s33, qV)
	}
}

// sortSingularValues sorts the singular values and adjusts V accordingly.
func sortSingularValues(
	b11, b12, b13,
	b21, b22, b23,
	b31, b32, b33 *float32,
	v11, v12, v13,
	v21, v22, v23,
	v31, v32, v33 *float32,
) {
	rho1 := dist2(*b11, *b21, *b31)
	rho2 := dist2(*b12, *b22, *b32)
	rho3 := dist2(*b13, *b23, *b33)
	c := rho1 < rho2
	condNegSwap(c, b11, b12)
	condNegSwap(c, v11, v12)
	condNegSwap(c, b21, b22)
	condNegSwap(c, v21, v22)
	condNegSwap(c, b31, b32)
	condNegSwap(c, v31, v32)
	condSwap(c, &rho1, &rho2)
	c = rho1 < rho3
	condNegSwap(c, b11, b13)
	condNegSwap(c, v11, v13)
	condNegSwap(c, b21, b23)
	condNegSwap(c, v21, v23)
	condNegSwap(c, b31, b33)
	condNegSwap(c, v31, v33)
	condSwap(c, &rho1, &rho3)
	c = rho2 < rho3
	condNegSwap(c, b12, b13)
	condNegSwap(c, v12, v13)
	condNegSwap(c, b22, b23)
	condNegSwap(c, v22, v23)
	condNegSwap(c, b32, b33)
	condNegSwap(c, v32, v33)
}

// QRGivensQuaternion computes the Givens rotation for QR decomposition.
func QRGivensQuaternion(a1, a2 float32, ch, sh *float32) {
	eps := float32(epsilon)
	rho := accurateSqrt(a1*a1 + a2*a2)

	if rho > eps {
		*sh = a2
	} else {
		*sh = 0
	}
	*ch = float32(math.Abs(float32(a1))) + float32(math.Max(float32(rho), float32(eps)))
	b := a1 < 0
	condSwap(b, sh, ch)
	w := rsqrt(*ch**ch + *sh**sh)
	*ch *= w
	*sh *= w
}

// QRDecomposition performs QR decomposition of a 3x3 matrix.
func QRDecomposition(
	b11, b12, b13, b21, b22, b23, b31, b32, b33 float32,
	q11, q12, q13,
	q21, q22, q23,
	q31, q32, q33 *float32,
	r11, r12, r13,
	r21, r22, r23,
	r31, r32, r33 *float32,
) {
	var ch1, sh1, ch2, sh2, ch3, sh3 float32
	var a, b float32

	// First Givens rotation
	QRGivensQuaternion(b11, b21, &ch1, &sh1)
	a = 1 - 2*sh1*sh1
	b = 2 * ch1 * sh1
	*r11 = a*b11 + b*b21
	*r12 = a*b12 + b*b22
	*r13 = a*b13 + b*b23
	*r21 = -b*b11 + a*b21
	*r22 = -b*b12 + a*b22
	*r23 = -b*b13 + a*b23
	*r31 = b31
	*r32 = b32
	*r33 = b33

	// Second Givens rotation
	QRGivensQuaternion(*r11, *r31, &ch2, &sh2)
	a = 1 - 2*sh2*sh2
	b = 2 * ch2 * sh2
	b11 = a**r11 + b**r31
	b12 = a**r12 + b**r32
	b13 = a**r13 + b**r33
	b21 = *r21
	b22 = *r22
	b23 = *r23
	b31 = -b**r11 + a**r31
	b32 = -b**r12 + a**r32
	b33 = -b**r13 + a**r33

	// Third Givens rotation
	QRGivensQuaternion(b22, b32, &ch3, &sh3)
	a = 1 - 2*sh3*sh3
	b = 2 * ch3 * sh3
	*r11 = b11
	*r12 = b12
	*r13 = b13
	*r21 = a*b21 + b*b31
	*r22 = a*b22 + b*b32
	*r23 = a*b23 + b*b33
	*r31 = -b*b21 + a*b31
	*r32 = -b*b22 + a*b32
	*r33 = -b*b23 + a*b33

	// Construct cumulative rotation Q = Q1 * Q2 * Q3
	sh12 := sh1 * sh1
	sh22 := sh2 * sh2
	sh32 := sh3 * sh3

	*q11 = (-1 + 2*sh12) * (-1 + 2*sh22)
	*q12 = 4*ch2*ch3*(-1+2*sh12)*sh2*sh3 + 2*ch1*sh1*(-1+2*sh32)
	*q13 = 4*ch1*ch3*sh1*sh3 - 2*ch2*(-1+2*sh12)*sh2*(-1+2*sh32)

	*q21 = 2 * ch1 * sh1 * (1 - 2*sh22)
	*q22 = -8*ch1*ch2*ch3*sh1*sh2*sh3 + (-1+2*sh12)*(-1+2*sh32)
	*q23 = -2*ch3*sh3 + 4*sh1*(ch3*sh1*sh3+ch1*ch2*sh2*(-1+2*sh32))

	*q31 = 2 * ch2 * sh2
	*q32 = 2 * ch3 * (1 - 2*sh22) * sh3
	*q33 = (-1 + 2*sh22) * (-1 + 2*sh32)
}

// SVD performs singular value decomposition on a 3x3 matrix.
func SVD(am ms3.Mat3) (U, S, V ms3.Mat3) {
	// Extract elements of A
	a := *(*Mat3)(unsafe.Pointer(&am))
	a11, a12, a13 := a.x00, a.x01, a.x02
	a21, a22, a23 := a.x10, a.x11, a.x12
	a31, a32, a33 := a.x20, a.x21, a.x22

	var u11, u12, u13 float32
	var u21, u22, u23 float32
	var u31, u32, u33 float32

	var s11, s12, s13 float32
	var s21, s22, s23 float32
	var s31, s32, s33 float32

	var v11, v12, v13 float32
	var v21, v22, v23 float32
	var v31, v32, v33 float32

	// Normal equations matrix
	var ATA11, ATA12, ATA13 float32
	var ATA21, ATA22, ATA23 float32
	var ATA31, ATA32, ATA33 float32

	multAtB(a11, a12, a13, a21, a22, a23, a31, a32, a33,
		a11, a12, a13, a21, a22, a23, a31, a32, a33,
		&ATA11, &ATA12, &ATA13, &ATA21, &ATA22, &ATA23, &ATA31, &ATA32, &ATA33)

	// Symmetric eigenanalysis
	var qV [4]float32
	jacobiEigenanalysis(&ATA11, &ATA21, &ATA22, &ATA31, &ATA32, &ATA33, &qV)
	quatToMat3(qV, &v11, &v12, &v13, &v21, &v22, &v23, &v31, &v32, &v33)

	// Compute B = A * V
	var b11, b12, b13 float32
	var b21, b22, b23 float32
	var b31, b32, b33 float32
	multAB(a11, a12, a13, a21, a22, a23, a31, a32, a33,
		v11, v12, v13, v21, v22, v23, v31, v32, v33,
		&b11, &b12, &b13, &b21, &b22, &b23, &b31, &b32, &b33)

	// Sort singular values and adjust V
	sortSingularValues(&b11, &b12, &b13, &b21, &b22, &b23, &b31, &b32, &b33,
		&v11, &v12, &v13, &v21, &v22, &v23, &v31, &v32, &v33)

	// QR decomposition to compute U and S
	QRDecomposition(b11, b12, b13, b21, b22, b23, b31, b32, b33,
		&u11, &u12, &u13, &u21, &u22, &u23, &u31, &u32, &u33,
		&s11, &s12, &s13, &s21, &s22, &s23, &s31, &s32, &s33)

	// Construct U, S, and V matrices
	U = convMat(Mat3{
		x00: u11, x01: u12, x02: u13,
		x10: u21, x11: u22, x12: u23,
		x20: u31, x21: u32, x22: u33,
	})
	S = convMat(Mat3{
		x00: s11, x01: s12, x02: s13,
		x10: s21, x11: s22, x12: s23,
		x20: s31, x21: s32, x22: s33,
	})
	V = convMat(Mat3{
		x00: v11, x01: v12, x02: v13,
		x10: v21, x11: v22, x12: v23,
		x20: v31, x21: v32, x22: v33,
	})
	return
}

// Mat3 is a 3x3 matrix.
type Mat3 struct {
	x00, x01, x02 float32
	x10, x11, x12 float32
	x20, x21, x22 float32
	// Padding to align to 16 bytes.
	_, _, _ float32
}

func convMat(a Mat3) ms3.Mat3 {
	return *(*ms3.Mat3)(unsafe.Pointer(&a))
}
