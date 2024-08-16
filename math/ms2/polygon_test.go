package ms2

import (
	"testing"

	math "github.com/chewxy/math32"
)

func TestPolygon_circle_smoothing(t *testing.T) {
	const r = 1
	const smooth = r
	var poly PolygonBuilder
	for _, facets := range []int{2, 3, 4, 6, 7} {
		poly.Reset()

		poly.AddXY(r, 0)
		poly.AddXY(r, r).Smooth(smooth, facets)
		poly.AddXY(0, r)
		poly.AddXY(-r, r).Smooth(smooth, facets)
		poly.AddXY(-r, 0)
		poly.AddXY(-r, -r).Smooth(smooth, facets)
		poly.AddXY(0, -r)
		poly.AddXY(r, -r).Smooth(smooth, facets)

		verts, err := poly.AppendVecs(nil)
		if err != nil {
			t.Fatal(err)
		}

		const tol = 1e-4
		wantVerts := 4 + (facets-1)*4
		if len(verts) != wantVerts {
			t.Errorf("want %d vertices, got %d", wantVerts, len(verts))
		}
		for _, v := range verts {
			gotR := Norm(v)
			if math.Abs(gotR-r) > tol {
				t.Error(v)
			}
		}
	}
}