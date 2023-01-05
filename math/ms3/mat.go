package ms3

import (
	"fmt"

	math "github.com/chewxy/math32"
)

const (
	badDim = "bad matrix dimensions"
	badIdx = "bad matrix index"
)

type Mat [3 * 3]float32

func NewMat(val []float32) *Mat {
	if len(val) != 9 {
		if val == nil {
			return &Mat{}
		}
		panic(badDim)
	}
	M := Mat{}
	copy(M[:], val)
	return &M
}

// Identity returns the Identity 3x3 matrix
func Identity() *Mat { return &Mat{1, 0, 0, 0, 1, 0, 0, 0, 1} }

func (m Mat) String() string {
	for i := range m {
		if math.Abs(m[i]) < 1e-6 {
			m[i] = 0
		}
	}
	return fmt.Sprintf("[%0.6g\t%0.6g\t%0.6g]\n[%0.6g\t%0.6g\t%0.6g]\n[%0.6g\t%0.6g\t%0.6g]",
		m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8])
}

func (m *Mat) Dims() (r, c int) { return 3, 3 }

func (m *Mat) At(i, j int) float32 {
	if uint(i) > 2 || uint(j) > 2 {
		panic(badIdx)
	}
	return m.at(i, j)
}

func (m *Mat) at(i, j int) float32 {
	return m[i*3+j]
}

func (m *Mat) Set(i, j int, v float32) {
	if uint(i) > 2 || uint(j) > 2 {
		panic(badIdx)
	}
	m.set(i, j, v)
}

// set sets element at position i,j without bounds checking
func (m *Mat) set(i, j int, v float32) {
	m[i*3+j] = v
}

// Scale multiplies the elements of a by f, placing the result in the receiver.
//
// See the Scaler interface for more information.
func (m *Mat) Scale(f float32, a *Mat) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.set(i, j, f*a.At(i, j))
		}
	}
}

// Performs matrix multiplication on v:
//
//	result = M * v
func (m *Mat) MulVec(v Vec) (result Vec) {
	result.X = v.X*m[0] + v.Y*m[1] + v.Z*m[2]
	result.Y = v.X*m[3] + v.Y*m[4] + v.Z*m[5]
	result.Z = v.X*m[6] + v.Y*m[7] + v.Z*m[8]
	return result
}

// Prod performs vector multiplication as if they were matrices
//
//	m = v1 * v2ᵀ
func (m *Mat) Prod(v1, v2t Vec) {
	m.Set(0, 0, v1.X*v2t.X)
	m.Set(0, 1, v1.X*v2t.Y)
	m.Set(0, 2, v1.X*v2t.Z)

	m.Set(1, 0, v1.Y*v2t.X)
	m.Set(1, 1, v1.Y*v2t.Y)
	m.Set(1, 2, v1.Y*v2t.Z)

	m.Set(2, 0, v1.Z*v2t.X)
	m.Set(2, 1, v1.Z*v2t.Y)
	m.Set(2, 2, v1.Z*v2t.Z)
}

// RotationFromQuat stores the rotation matrix from a quaternion (must be normalized).
//
//	[R] = q.Real*q.Real * [E] - dot([q],[q])*[E] + 2*[q]*[q]ᵀ + 2*q.Real * skew([q])
//
// where
//
//	[q] = {q.Imag, q.Jmag, q.Kmag}
//
// and
//
//	[E]=Eye()
func (m *Mat) RotationFromQuat(q Quat) {
	qv := q.V
	var qs Mat
	qs.Skew(qv)
	q01 := Identity()
	q01.Scale(q.W*q.W, q01)

	qd := Identity()
	qd.Scale(Dot(qv, qv), qd)

	qs.Scale(2*q.W, &qs)

	qv = Scale(2, qv)
	m.Prod(qv, qv) // m = 2*[q]*[q]ᵀ
	m.Add(m, q01)  // m += q.Real*q.Real * [E]
	m.Add(m, qd)   // m += dot([q],[q])*[E]
	m.Add(m, &qs)  // m += 2*q.Real * skew([q])
}

// Performs transposed matrix multiplication on v:
//
//	result = Mᵀ * v
func (m *Mat) MulVecTrans(v Vec) (result Vec) {
	result.X = v.X*m[0] + v.Y*m[3] + v.Z*m[6]
	result.Y = v.X*m[1] + v.Y*m[4] + v.Z*m[7]
	result.Z = v.X*m[2] + v.Y*m[5] + v.Z*m[8]
	return result
}

// Skew returns the skew symmetric matrix (right hand system) of v.
//
//	                [0  -z  y]
//	Skew({x,y,z}) = [z   0 -x]
//	                [-y  x  0]
func (m *Mat) Skew(v Vec) {
	*m = Mat{
		0, -v.Z, v.Y,
		v.Z, 0, -v.X,
		-v.Y, v.X, 0,
	}
}

// Mul takes the matrix product of a and b, placing the result in the receiver.
// If the number of columns in a does not equal 3, Mul will panic.
func (m *Mat) Mul(a, b *Mat) {
	m.mul(a, b)
}

// Subtracts b from a and stores result in receiver. Sub will panic if the two matrices do not have the same shape.
func (m *Mat) Sub(a, b *Mat) {
	for i := range m {
		m[i] = a[i] - b[i]
	}
}

// Add adds a and b element-wise, placing the result in the receiver. Add will panic if the two matrices do not have the same shape.
func (m *Mat) Add(A, B *Mat) {
	for i := range m {
		m[i] = A[i] + B[i]
	}
}

func (m *Mat) VecRow(i int) Vec {
	if uint(i) > 2 {
		panic(badIdx)
	}
	return Vec{X: m.at(i, 0), Y: m.at(i, 1), Z: m.at(i, 2)}
}

func (m *Mat) VecCol(j int) Vec {
	if uint(j) > 2 {
		panic(badIdx)
	}
	return Vec{X: m.at(0, j), Y: m.at(1, j), Z: m.at(2, j)}
}

func (C *Mat) mul(A, B *Mat) {
	for i := 0; i < 3; i++ {
		ridx := i * 3
		for j := 0; j < 3; j++ {
			var tmp float32
			for k := 0; k < 3; k++ {
				tmp += A.At(i, k) * B.At(k, j)
			}
			C[ridx+j] = tmp
		}
	}
}
