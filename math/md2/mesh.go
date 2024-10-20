// DO NOT EDIT.
// This file was generated automatically
// from gen.go. Please do not edit this file.

package md2

// AppendGrid splits the argument bounds [Box] x,y axes by nx,ny, respectively
// and generates points on the vertices generated by the division and appends them to dst, returning the result.
//
// Indexing is x-major:
//
//	grid := ms2.AppendGrid(nil, bb, nx, ny)
//	ix, iy := 1, 0
//	pos := grid[iy*nx + ix]
func AppendGrid(dst []Vec, bounds Box, nx, ny int) []Vec {
	if nx <= 0 || ny <= 0 {
		panic("bad AppendGrid argument")
	}
	nxyz := Vec{X: float64(nx - 1), Y: float64(ny - 1)}
	dxyz := DivElem(bounds.Size(), nxyz)
	var xyz Vec
	for j := 0; j < ny; j++ {
		xyz.Y = bounds.Min.Y + dxyz.Y*float64(j)
		for i := 0; i < nx; i++ {
			xyz.X = bounds.Min.X + dxyz.X*float64(i)
			dst = append(dst, xyz)
		}
	}
	return dst
}
