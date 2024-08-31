package ms2

import (
	"errors"

	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms1"
)

// PolygonBuilder facilitates polygon construction with arcs, smoothing and chamfers
// with the [PolygonVertex] type.
type PolygonBuilder struct {
	verts []PolygonVertex
}

// PolygonVertex represents a polygon point joined by two edges. It is
// used by the [PolygonBuilder] type.
type PolygonVertex struct {
	v      Vec     // Absolute vertex position.
	radius float32 // Smoothing radius, if zero then no smoothing.
	facets int32   // Amount of facets to create when smoothing. If negative indicates arcing instead of smoothing.
}

// Nagon sets the vertices of p to that of a N sided regular polygon. If n<3 then Nagon does nothing.
func (p *PolygonBuilder) Nagon(n int, centerDistance float32) {
	p.NagonSmoothed(n, centerDistance, 0, 0)
}

// NagonSmoothed sets the vertices of p to that of a N sided regular polygon and smoothes the result.
// If n<3 or radius<centerDistance then Nagon does nothing.
func (p *PolygonBuilder) NagonSmoothed(n int, centerDistance float32, facets int, radius float32) {
	if n < 3 || (radius != 0 && radius > centerDistance) {
		return
	}
	p.verts = p.verts[:0] // Reset buffer.
	m := RotationMat2(2 * math.Pi / float32(n))
	v := Vec{X: centerDistance, Y: 0}
	for i := 0; i < n; i++ {
		p.Add(v).Smooth(radius, facets)
		v = MulMatVec(m, v)
	}
}

// Add adds a point in absolute cartesian coordinates to the polygon being built.
func (p *PolygonBuilder) Add(v Vec) *PolygonVertex {
	p.verts = append(p.verts, PolygonVertex{v: v})
	return &p.verts[len(p.verts)-1]
}

// AddXY adds a point in absolute cartesian coordinates to the polygon being built.
func (p *PolygonBuilder) AddXY(x, y float32) *PolygonVertex {
	return p.Add(Vec{X: x, Y: y})
}

// AddPolarRTheta adds a point in absolute polar coordinates to the polygon being built.
func (p *PolygonBuilder) AddPolarRTheta(r, theta float32) *PolygonVertex {
	return p.Add(pol{R: r, Theta: theta}.Cartesian())
}

// AddRelative adds a point in absolute cartesian coordinates to the polygon being built, relative to last vertex added.
// If no vertices present then takes origin (x=0,y=0) as reference.
func (p *PolygonBuilder) AddRelative(v Vec) *PolygonVertex {
	last := p.last()
	if last == nil {
		return p.Add(v) // If no vertices present take origin as start point.
	}
	return p.Add(Add(last.v, v))
}

// AddRelativeXY is shorthand for [PolygonBuilder.AddRelative]([Vec]{x,y}).
func (p *PolygonBuilder) AddRelativeXY(x, y float32) *PolygonVertex {
	return p.AddRelative(Vec{X: x, Y: y})
}

// DropLast drops the last vertex. Can be called multiple times to drop several vertices.
func (p *PolygonBuilder) DropLast() {
	if len(p.verts) > 0 {
		p.verts = p.verts[:len(p.verts)-1]
	}
}

// Reset resets all polygon builder state dropping all vertices.
func (p *PolygonBuilder) Reset() {
	p.verts = p.verts[:0]
}

// AppendVecs appends the Polygon's discretized representation to the argument Vec buffer and returns the result.
// It does not change the internal state of the PolygonBuilder and thus can be called repeatedly.
func (p *PolygonBuilder) AppendVecs(buf []Vec) ([]Vec, error) {
	if len(p.verts) < 2 {
		return buf, errors.New("too few vertices")
	}
	prev := p.verts[len(p.verts)-1]
	for i := range p.verts {
		current := p.verts[i]
		if current.isArc() {
			buf = appendArc(buf, current, prev)
			buf = append(buf, current.v)
		} else if current.isSmoothed() {
			buf = append(buf, current.v)
			buf = buf[:len(buf)-1] // Smoothed vertex is consumed and replaced.
			next := p.verts[(i+1)%len(p.verts)]
			buf = appendSmooth(buf, current, prev, next)
		} else {
			buf = append(buf, current.v)
		}
		prev = current
	}
	return buf, nil
}

func (p *PolygonBuilder) last() *PolygonVertex {
	if len(p.verts) > 0 {
		return &p.verts[len(p.verts)-1]
	}
	return nil
}

// Smooth smoothes this polygon vertex by a radius and discretises the smoothing in facets.
func (v *PolygonVertex) Smooth(radius float32, facets int) {
	if radius > 0 && facets > 0 {
		v.radius = radius
		v.facets = int32(facets)
	}
}

// Arc creates an arc between this and the previous PolygonVertex
// discretised by facets starting at previous radius.
//
// A positive radius specifies counter-clockwise path,
// a negative radius specifies a clockwise path.
func (v *PolygonVertex) Arc(radius float32, facets int) {
	if radius != 0 && facets > 0 {
		v.radius = radius
		v.facets = -int32(facets)
	}
}
func (v *PolygonVertex) isSmoothed() bool { return v.facets > 0 && v.radius > 0 }
func (v *PolygonVertex) isArc() bool      { return v.facets < 0 && v.radius != 0 }

const sqrtHalf = math.Sqrt2 / 2

// Chamfer is a smoothing of a single facet of length `size`.
func (v *PolygonVertex) Chamfer(size float32) {
	v.Smooth(size*sqrtHalf, 1)
}

func appendArc(buf []Vec, vCurr, vPrev PolygonVertex) []Vec {
	facets := -vCurr.facets
	if facets == 1 {
		return buf
	}
	r := vCurr.radius
	return appendArc2points(buf, vPrev.v, vCurr.v, r, facets)
}

func appendArc2points(dst []Vec, p1, p2 Vec, r float32, facets int32) []Vec {
	isNeg := r < 0
	if isNeg {
		r = -r
	}
	V12 := Sub(p2, p1)
	chordCenter := Add(p1, Scale(0.5, V12))
	chordLen := Norm(V12) // Chord length.
	if chordLen > 2*r {
		return dst
	}
	// Theta is the opening angle from the center of the arc circle
	// to the two chord points.
	// Due to chord definition theta/2 is the angle formed
	// by the chord and the tangent to the chord point.
	chordThetaDiv2 := math.Asin(chordLen / (2 * r))
	diffTo90 := chordThetaDiv2 - math.Pi/2
	if math.Abs(diffTo90) < 1e-6 {
		// Ill conditioned arc. Do a little correction away from the 90 degree mark.
		chordThetaDiv2 += math.Copysign(1e-6, -diffTo90)
	}
	// dtheta will be the angle between points on the arc.
	dtheta := 2 * chordThetaDiv2 / float32(facets)

	// We now find the arc center. To do this we look at the radius
	// sign which will tell us the orientation of the arc (clockwise vs counterclockwise).
	var perp Vec
	if isNeg {
		// Y:-X -> for the simple case V12=Vec{X:1} this results in
		// perp=Vec{Y:-1}. If we place arc center in that direction
		// that will result in a clockwise arc, so negative angle.
		dtheta = -dtheta
		perp = Vec{X: V12.Y, Y: -V12.X}
	} else {
		perp = Vec{X: -V12.Y, Y: V12.X}
	}
	// x is distance from arc center to chord center.
	x := 0.5 * chordLen / math.Tan(chordThetaDiv2)
	// Perp is scaled to be of length x.
	// Then, simply add perp for arc center.
	perp = Scale(x/chordLen, perp)
	arcCenter := Add(chordCenter, perp)
	// Begin from p1 and make way to point right before p2. We do not add p1 and p2.
	T := RotationMat2(dtheta)
	rv := Sub(p1, arcCenter)
	for i := int32(0); i < facets-1; i++ {
		rv = MulMatVec(T, rv)
		dst = append(dst, Add(arcCenter, rv))
	}
	return dst
}

func appendSmooth(buf []Vec, v, vPrev, vNext PolygonVertex) []Vec {
	if !v.isSmoothed() || v.facets == 1 {
		return buf
	}
	r := v.radius
	facets := v.facets

	r = math.Abs(r)
	// Work out angle.
	vp := Sub(vPrev.v, v.v)
	vn := Sub(vNext.v, v.v)
	normVP := Norm(vp)
	normVN := Norm(vn)
	// l1 := Line{vPrev.v, v.v}
	// p1 := l1.Interpolate((normVP - r) / normVP)
	// l2 := Line{vNext.v, v.v}
	// p2 := l2.Interpolate((normVN - r) / normVN)
	// p3 :=
	v0 := Scale(1./normVP, vp)
	v1 := Scale(1./normVN, vn)
	theta := math.Acos(Dot(vp, vn) / (normVN * normVP))

	d1 := r / math.Tan(theta/2)
	if d1 > normVP || d1 > normVN || math.IsNaN(theta) {
		return buf
	}

	p0 := Add(v.v, Scale(d1, v0)) // Tangent points.

	d2 := r / math.Sin(theta/2)

	vc := Unit(Add(v0, v1))
	c := Add(v.v, Scale(d2, vc))

	dtheta := ms1.Sign(Cross(v1, v0)) * (math.Pi - theta) / float32(facets)

	T := RotationMat2(dtheta) // rotation matrix.
	rv := Sub(p0, c)          // radius vector

	for i := int32(0); i < facets-1; i++ {
		rv = MulMatVec(T, rv)
		buf = append(buf, Add(c, rv))
	}
	return buf
}
