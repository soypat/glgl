package ms3

import (
	math "github.com/chewxy/math32"
)

// Constants used in the algorithm
const (
	gamma   = 5.828427124  // FOUR_GAMMA_SQUARED = sqrt(8)+3
	cstar   = 0.923879532  // cos(pi/8)
	sstar   = 0.3826834323 // sin(pi/8)
	epsilon = 1e-6
)

// SVD performs singular value decomposition on a 3x3 matrix.
func (a Mat3) SVD() (U, S, V Mat3) {
	// Extract elements of A
	// Normal equations matrix
	ATA := MulMat3(a.Transpose(), a)
	// Symmetric eigenanalysis
	var qV [4]float32
	jacobiEigenanalysis(&ATA.x00, &ATA.x10, &ATA.x11, &ATA.x20, &ATA.x21, &ATA.x22, &qV)

	// Compute B = A * V
	V = RotationMat3(Quat{I: qV[0], J: qV[1], K: qV[2], W: qV[3]})
	b := MulMat3(a, V)
	// Sort singular values and adjust V
	b, V = sortSingularValues(b, V)

	// QR decomposition to compute U and S
	U, S = b.QRDecomposition()
	return U, S, V
}

// QRDecomposition performs QR decomposition of a 3x3 matrix using Mat3 type.
func (b Mat3) QRDecomposition() (q, r Mat3) {
	// Extract elements from bb
	b11, b12, b13 := b.x00, b.x01, b.x02
	b21, b22, b23 := b.x10, b.x11, b.x12
	b31, b32, b33 := b.x20, b.x21, b.x22

	// First Givens rotation
	ch1, sh1 := qrGivensQuat(b11, b21)
	as := 1 - 2*sh1*sh1
	bs := 2 * ch1 * sh1

	// Compute R after first rotation
	r = Mat3{
		x00: as*b11 + bs*b21,
		x01: as*b12 + bs*b22,
		x02: as*b13 + bs*b23,
		x10: -bs*b11 + as*b21,
		x11: -bs*b12 + as*b22,
		x12: -bs*b13 + as*b23,
		x20: b31,
		x21: b32,
		x22: b33,
	}

	// Second Givens rotation
	ch2, sh2 := qrGivensQuat(r.x00, r.x20)
	as = 1 - 2*sh2*sh2
	bs = 2 * ch2 * sh2

	b11 = as*r.x00 + bs*r.x20
	b12 = as*r.x01 + bs*r.x21
	b13 = as*r.x02 + bs*r.x22
	b21 = r.x10
	b22 = r.x11
	b23 = r.x12
	b31 = -bs*r.x00 + as*r.x20
	b32 = -bs*r.x01 + as*r.x21
	b33 = -bs*r.x02 + as*r.x22

	// Third Givens rotation
	ch3, sh3 := qrGivensQuat(b22, b32)
	as = 1 - 2*sh3*sh3
	bs = 2 * ch3 * sh3

	// Compute R after third rotation
	r = Mat3{
		x00: b11,
		x01: b12,
		x02: b13,
		x10: as*b21 + bs*b31,
		x11: as*b22 + bs*b32,
		x12: as*b23 + bs*b33,
		x20: -bs*b21 + as*b31,
		x21: -bs*b22 + as*b32,
		x22: -bs*b23 + as*b33,
	}

	// Construct cumulative rotation Q = Q1 * Q2 * Q3
	sh12 := sh1 * sh1
	sh22 := sh2 * sh2
	sh32 := sh3 * sh3

	q = Mat3{
		x00: (-1 + 2*sh12) * (-1 + 2*sh22),
		x01: 4*ch2*ch3*(-1+2*sh12)*sh2*sh3 + 2*ch1*sh1*(-1+2*sh32),
		x02: 4*ch1*ch3*sh1*sh3 - 2*ch2*(-1+2*sh12)*sh2*(-1+2*sh32),

		x10: 2 * ch1 * sh1 * (1 - 2*sh22),
		x11: -8*ch1*ch2*ch3*sh1*sh2*sh3 + (-1+2*sh12)*(-1+2*sh32),
		x12: -2*ch3*sh3 + 4*sh1*(ch3*sh1*sh3+ch1*ch2*sh2*(-1+2*sh32)),

		x20: 2 * ch2 * sh2,
		x21: 2 * ch3 * (1 - 2*sh22) * sh3,
		x22: (-1 + 2*sh22) * (-1 + 2*sh32),
	}
	return q, r
}

// sortSingularValues sorts the singular values and adjusts V accordingly.
func sortSingularValues(b, v Mat3) (Mat3, Mat3) {
	rho1 := Norm2(b.VecCol(0))
	rho2 := Norm2(b.VecCol(1))
	rho3 := Norm2(b.VecCol(2))
	if rho1 < rho2 {
		b.x00, b.x01 = b.x01, -b.x00
		v.x00, v.x01 = v.x01, -v.x00

		b.x10, b.x11 = b.x11, -b.x10
		v.x10, v.x11 = v.x11, -v.x10

		b.x20, b.x21 = b.x21, -b.x20
		v.x20, v.x21 = v.x21, -v.x20
		rho1, rho2 = rho2, rho1
	}

	if rho1 < rho3 {
		b.x00, b.x02 = b.x02, -b.x00
		v.x00, v.x02 = v.x02, -v.x00

		b.x10, b.x12 = b.x12, -b.x10
		v.x10, v.x12 = v.x12, -v.x10

		b.x20, b.x22 = b.x22, -b.x20
		v.x20, v.x22 = v.x22, -v.x20
		rho3 = rho1 // no need to assign rho1 here.
	}

	if rho2 < rho3 {
		b.x01, b.x02 = b.x02, -b.x01
		v.x01, v.x02 = v.x02, -v.x01

		b.x11, b.x12 = b.x12, -b.x11
		v.x11, v.x12 = v.x12, -v.x11

		b.x21, b.x22 = b.x22, -b.x21
		v.x21, v.x22 = v.x22, -v.x21
	}
	return b, v
}

// qrGivensQuat computes the Givens rotation for QR decomposition.
func qrGivensQuat(a1, a2 float32) (ch, sh float32) {
	eps := float32(epsilon)
	rho := accurateSqrt(a1*a1 + a2*a2)

	if rho > eps {
		sh = a2
	} else {
		sh = 0
	}
	ch = math.Abs(a1) + math.Max(rho, eps)
	w := rsqrt(ch*ch + sh*sh)
	ch *= w
	sh *= w
	if a1 < 0 {
		sh, ch = ch, sh
	}
	return ch, sh
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

// jacobiConjugation performs the Jacobi rotation to diagonalize the matrix.
func jacobiConjugation(x, y, z int,
	s11 *float32,
	s21, s22 *float32,
	s31, s32, s33 *float32,
	qV *[4]float32) {
	ch, sh := approximateGivensQuaternion(*s11, *s21, *s22)

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

// approximateGivensQuaternion computes the Givens rotation quaternion.
func approximateGivensQuaternion(a11, a12, a22 float32) (ch, sh float32) {
	ch = 2 * (a11 - a22)
	sh = a12
	b := gamma*sh*sh < ch*ch
	w := rsqrt(ch*ch + sh*sh)
	if b {
		ch = w * ch
		sh = w * sh
	} else {
		ch = cstar
		sh = sstar
	}
	return ch, sh
}

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
	// return math32.Sqrt(x)
	return x * rsqrt1(x)
}
