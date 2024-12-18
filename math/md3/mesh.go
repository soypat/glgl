// DO NOT EDIT.
// This file was generated automatically
// from gen.go. Please do not edit this file.

package md3

// AppendGrid splits the argument bounds [Box] x,y,z axes by nx,ny,nz, respectively
// and generates points on the vertices generated by the division and appends them to dst, returning the result.
//
// Indexing is x-major, y-second-major:
//
//	grid := ms3.AppendGrid(nil, bb, nx, ny, nz)
//	ix, iy, iz := 1, 0, 3
//	pos := grid[iz*(nx+ny) + iy*nx + ix]
func AppendGrid(dst []Vec, bounds Box, nx, ny, nz int) []Vec {
	if nx <= 0 || ny <= 0 || nz <= 0 {
		panic("bad AppendGrid argument")
	}
	nxyz := Vec{X: float64(nx - 1), Y: float64(ny - 1), Z: float64(nz - 1)}
	dxyz := DivElem(bounds.Size(), nxyz)
	var xyz Vec
	for k := 0; k < nz; k++ {
		xyz.Z = bounds.Min.Z + dxyz.Z*float64(k)
		for j := 0; j < ny; j++ {
			xyz.Y = bounds.Min.Y + dxyz.Y*float64(j)
			for i := 0; i < nx; i++ {
				xyz.X = bounds.Min.X + dxyz.X*float64(i)
				dst = append(dst, xyz)
			}
		}
	}
	return dst
}