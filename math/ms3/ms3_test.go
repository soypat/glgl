package ms3

import (
	"math"
	"testing"
)

func TestRotation(t *testing.T) {
	const tol = 1e-7
	v := Vec{X: 1}
	y90 := RotationMat4(math.Pi/2, Vec{Y: 1})
	got := y90.MulPosition(v)
	want := Vec{Z: -1}
	if !EqualElem(got, want, tol) {
		t.Errorf("want %v, got %v", want, got)
	}
}
