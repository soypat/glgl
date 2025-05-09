// DO NOT EDIT.
// This file was generated automatically
// from gen.go. Please do not edit this file.

package md2

import (
	math "math"
)

// AppendGrid splits the argument bounds [Box] x,y axes by nx,ny, respectively
// and generates points on the vertices generated by the division and appends them to dst, returning the result.
// All box edges are vertices in result. AppendGrid panics if it receives a dimension less than 2.
//
// Indexing is x-major:
//
//	grid := ms2.AppendGrid(nil, domain, nx, ny)
//	ix, iy := 1, 0
//	pos := grid[iy*nx + ix]
//
// Deprecated: Maintenance of glgl math packages is moving to https://github.com/soypat/geometry.
func AppendGrid(dst []Vec, domain Box, nx, ny int) []Vec {
	if nx <= 1 || ny <= 1 {
		panic("AppendGrid needs more grid subdivisions")
	}
	nxyz := Vec{X: float64(nx - 1), Y: float64(ny - 1)}
	dxyz := DivElem(domain.Size(), nxyz)
	var xyz Vec
	for j := 0; j < ny; j++ {
		xyz.Y = domain.Min.Y + dxyz.Y*float64(j)
		for i := 0; i < nx; i++ {
			xyz.X = domain.Min.X + dxyz.X*float64(i)
			dst = append(dst, xyz)
		}
	}
	return dst
}

// GridSubdomain facilitates obtaining the set of points in a grid shared between a domain box
// and a subdomain box contained within the domain box. Points of the grid should
// be ordered in x-major format, like the values returned by [AppendGrid].
//
//	istart, nxSub, nySub := GridSubdomain(domain, nx, ny, subdomain)
//	for iy := 0; iy < nySub; iy++ {
//		off := istart + iy*nx
//		for ix := 0; ix < nxSub; ix++ {
//			pointInSubdomain := grid[off+ix]
//			// do something with pointInSubdomain.
//		}
//	}
func GridSubdomain(domain Box, nxDomain, nyDomain int, subdomain Box) (iStart, nxSub, nySub int) {
	if !domain.ContainsBox(subdomain) {
		panic("subdomain not contained in domain")
	}
	dx := (domain.Max.X - domain.Min.X) / float64(nxDomain-1)
	dy := (domain.Max.Y - domain.Min.Y) / float64(nyDomain-1)

	off := Sub(subdomain.Min, domain.Min)
	ix0 := iceil(off.X / dx)
	iy0 := iceil(off.Y / dy)
	iStart = ix0 + iy0*nxDomain

	offEnd := Sub(subdomain.Max, domain.Min)
	ixf := int(offEnd.X / dx)
	iyf := int(offEnd.Y / dy)

	nxSub = ixf - ix0 + 1
	nySub = iyf - iy0 + 1
	return iStart, nxSub, nySub
}

func iceil(f float64) int {
	return int(math.Ceil(f))
}
