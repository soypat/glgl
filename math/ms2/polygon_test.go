package ms2

import (
	"testing"

	math "github.com/chewxy/math32"
)

var testoffsets = []Vec{{-1, -2}, {-2, 1}, {2, -1}, {}, {1, 0}, {0, 1}, {1, 1}}

func TestPolygon_circle_smoothing(t *testing.T) {
	var poly PolygonBuilder
	for _, offset := range testoffsets {
		for _, radius := range []float32{1e-6, 0.1, 1, 5, 100} {
			r := radius
			smooth := radius
			for _, facets := range []int{2, 3, 4, 6, 7} {
				poly.Reset()

				poly.AddXY(offset.X+r, offset.Y+0)
				poly.AddXY(offset.X+r, offset.Y+r).Smooth(smooth, facets)
				poly.AddXY(offset.X+0, offset.Y+r)
				poly.AddXY(offset.X-r, offset.Y+r).Smooth(smooth, facets)
				poly.AddXY(offset.X-r, offset.Y+0)
				poly.AddXY(offset.X-r, offset.Y-r).Smooth(smooth, facets)
				poly.AddXY(offset.X+0, offset.Y-r)
				poly.AddXY(offset.X+r, offset.Y-r).Smooth(smooth, facets)

				verts, err := poly.AppendVecs(nil)
				if err != nil {
					t.Fatal(err, offset, r)
				}

				const tol = 1e-4
				wantVerts := 4 + (facets-1)*4
				if len(verts) != wantVerts {
					t.Errorf("want %d vertices, got %d", wantVerts, len(verts))
				}
				for _, v := range verts {
					vr := Sub(v, offset)
					gotR := Norm(vr)
					if math.IsNaN(gotR) {
						t.Error("NaN", v)
					}
					diff := gotR - r
					if math.Abs(diff) > tol || math.IsNaN(diff) {
						t.Error(v)
					}
				}
			}
		}
	}
}

func TestPolygon_circle_arcing(t *testing.T) {
	var poly PolygonBuilder
	for _, offset := range testoffsets {
		for _, radius := range []float32{0.1, 1, 5, 100} {
			r := radius
			for _, facets := range []int{2, 3, 4, 6, 7} {
				poly.Reset()
				poly.AddXY(offset.X+r, offset.Y).Arc(r, facets)
				poly.AddXY(offset.X-r, offset.Y).Arc(r, facets)

				verts, err := poly.AppendVecs(nil)
				if err != nil {
					t.Fatal(err, offset, radius)
				}

				const tol = 1e-4
				wantVerts := 2 + (facets-1)*2
				if len(verts) != wantVerts {
					t.Errorf("want %d vertices, got %d", wantVerts, len(verts))
				}
				for _, v := range verts {
					vr := Sub(v, offset)
					gotR := Norm(vr)
					if math.Abs(gotR-r) > tol {
						t.Error(offset, v)
					}
				}
			}
		}
	}
}

func TestSmooth_radiusLimitBug(t *testing.T) {
	var poly PolygonBuilder
	poly.AddXY(0.8, 0.6)
	poly.AddXY(0.4, 0.6).Smooth(0.2, 5)
	poly.AddXY(0.3, 0)
	vecs, err := poly.AppendVecs(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(vecs)
}

func TestArc_radiusLimitBug(t *testing.T) {
	var poly PolygonBuilder
	poly.AddXY(1, 0)
	poly.AddXY(0, 38).Arc(380, 2)
	vecs, err := poly.AppendVecs(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(vecs)
}

func TestPolygon_IsClockwise(t *testing.T) {
	var tests = []struct {
		verts  []Vec
		wantCW bool
	}{
		{ // Counterclockwise triangle.
			verts:  []Vec{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}},
			wantCW: false,
		},
		{ // Clockwise triangle.
			verts:  []Vec{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
			wantCW: true,
		},
	}
	var poly PolygonBuilder
	for _, test := range tests {
		poly.Reset()
		for _, v := range test.verts {
			poly.Add(v)
		}
		gotCW := poly.IsClockwise()
		if test.wantCW != gotCW {
			t.Errorf("want CW=%v got CW=%v", test.wantCW, gotCW)
		}
	}
}
