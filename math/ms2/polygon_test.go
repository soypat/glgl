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
