package ms2

import (
	"math/rand"
	"testing"
)

func TestGridSubdomain(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	contained := make(map[int][2]int)
	var grid []Vec
	fails := 0
	pass := 0
	const maxDiv = 128
	for i := 0; i < 32; i++ {
		nx := rng.Intn(maxDiv) + 2
		ny := rng.Intn(maxDiv) + 2
		domain := randBox(randIVec(rng), rng)
		subdomain := randSubBox(domain, rng)

		grid = AppendGrid(grid[:0], domain, nx, ny)
		istart, nxSub, nySub := GridSubdomain(domain, nx, ny, subdomain)
		for iy := 0; iy < nySub; iy++ {
			off := istart + iy*nx
			for ix := 0; ix < nxSub; ix++ {
				contained[off+ix] = [2]int{ix, iy}
			}
		}
		for iy := 0; iy < ny; iy++ {
			off := iy * nx
			for ix := 0; ix < nx; ix++ {
				idx := off + ix
				p := grid[idx]
				subIdx, got := contained[idx]
				want := subdomain.Contains(p)
				if got != want {
					if !got {
						subIdx = [2]int{-1, -1}
					}
					fails++
					t.Logf("point OOB (ix,iy)=(%d, %d) (x,y)=(%.1f,%.1f) subdomain=%.1f wantContain=%v, gotContain=%v  subidx=(%d, %d)/%d", ix, iy, p.X, p.Y, subdomain, want, got, subIdx[0], subIdx[1], len(contained))
				} else {
					pass++
				}
			}
		}
		for k := range contained {
			delete(contained, k)
		}
	}
	fracPass := float64(pass) / (float64(pass + fails))
	t.Logf("passed %.2f%%", 100*fracPass)
	if fracPass < 0.995 {
		t.Errorf("too many failures")
	}

}

func randBox(min Vec, rng *rand.Rand) Box {
	return Box{
		Min: min,
		Max: Add(min, randIVec(rng)),
	}
}

func randIVec(rng *rand.Rand) Vec {
	nx, ny := rng.Intn(11)+1, rng.Intn(11)+1
	return Vec{X: float32(nx), Y: float32(ny)}
}

func randSubBox(domain Box, rng *rand.Rand) (sub Box) {
	sz := domain.Size()
	for sub.Empty() {
		newSz := DivElem(sz, randIVec(rng))
		off := DivElem(sz, randIVec(rng))
		sub = Box{
			Min: Add(domain.Min, off),
			Max: MinElem(domain.Max, Add(domain.Min, newSz)),
		}
	}

	if !domain.ContainsBox(sub) {
		panic("bad randSubBox implementation")
	}
	return sub
}

// subsz := subdomain.Size()
// const tol = 1e-3
// if ms1.EqualWithinAbs(domain.Min.X, subdomain.Min.X, tol) || domain.Min.X == subdomain.Min.X {
// 	subdomain.Min.X -= 1e-3 * subsz.X
// }
// if ms1.EqualWithinAbs(domain.Min.Y, subdomain.Min.Y, tol) || domain.Min.Y == subdomain.Min.Y {
// 	subdomain.Min.Y -= 1e-3 * subsz.Y
// }
// if ms1.EqualWithinAbs(domain.Max.X, subdomain.Max.X, tol) || domain.Max.X == subdomain.Max.X {
// 	subdomain.Max.X += 1e-3 * subsz.X
// }
// if ms1.EqualWithinAbs(domain.Max.Y, subdomain.Max.Y, tol) || domain.Max.Y == subdomain.Max.Y {
// 	subdomain.Max.Y += 1e-3 * subsz.Y
// }
