package ms2

import (
	"errors"
	"strconv"

	math "github.com/chewxy/math32"
)

type cpAtIdxErr struct {
	idx int
	msg error
}

func (cperr *cpAtIdxErr) Error() string {
	return "PolygonBuilder control_point[" + strconv.Itoa(cperr.idx) + "]: " + cperr.msg.Error()
}

const arcTol = 5e-1

// PolygonBuilder facilitates polygon construction with arcs, smoothing and chamfers
// with the [PolygonControlPoint] type.
type PolygonBuilder struct {
	verts []PolygonControlPoint
}

// PolygonControlPoint represents a polygon point joined by two edges, or alternatively
// a smoothed control point, in which case the vertex does not lie in the polygon.
// It is used by the [PolygonBuilder] type and notably returned by the Add* methods
// so that the user may control the polygon's shape. By default represents a vertex joining two other neighboring vertices.
type PolygonControlPoint struct {
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
	p.Reset()
	m := RotationMat2(2 * math.Pi / float32(n))
	v := Vec{X: centerDistance, Y: 0}
	for i := 0; i < n; i++ {
		p.Add(v).Smooth(radius, facets)
		v = MulMatVec(m, v)
	}
}

// Add adds a point in absolute cartesian coordinates to the polygon being built.
func (p *PolygonBuilder) Add(v Vec) *PolygonControlPoint {
	p.verts = append(p.verts, PolygonControlPoint{v: v})
	return &p.verts[len(p.verts)-1]
}

// AddXY adds a point in absolute cartesian coordinates to the polygon being built.
func (p *PolygonBuilder) AddXY(x, y float32) *PolygonControlPoint {
	return p.Add(Vec{X: x, Y: y})
}

// AddPolarRTheta adds a point in absolute polar coordinates to the polygon being built.
func (p *PolygonBuilder) AddPolarRTheta(r, theta float32) *PolygonControlPoint {
	return p.Add(pol{R: r, Theta: theta}.Cartesian())
}

// AddRelative adds a point in absolute cartesian coordinates to the polygon being built, relative to last vertex added.
// If no vertices present then takes origin (x=0,y=0) as reference.
func (p *PolygonBuilder) AddRelative(v Vec) *PolygonControlPoint {
	last := p.last()
	if last == nil {
		return p.Add(v) // If no vertices present take origin as start point.
	}
	return p.Add(Add(last.v, v))
}

// AddRelativeXY is shorthand for [PolygonBuilder.AddRelative]([Vec]{x,y}).
func (p *PolygonBuilder) AddRelativeXY(x, y float32) *PolygonControlPoint {
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
	var err error
	prev := p.verts[len(p.verts)-1]
	for i := range p.verts {
		current := p.verts[i]
		if i != 0 && prev == current {
			return buf, &cpAtIdxErr{idx: i, msg: errCPEqualToPrev}
		}
		if current.isArc() {
			buf, err = appendArc2points(buf, prev.v, current.v, current.radius, -current.facets)
			buf = append(buf, current.v)
		} else if current.isSmoothed() {
			next := p.verts[(i+1)%len(p.verts)]
			buf, err = appendSmoothedCorner(buf, prev.v, current.v, next.v, current.radius, current.facets)
		} else {
			buf = append(buf, current.v)
		}
		if err != nil {
			return buf, &cpAtIdxErr{idx: i, msg: err}
		}
		prev = current
	}
	return buf, nil
}

func (p *PolygonBuilder) last() *PolygonControlPoint {
	if len(p.verts) > 0 {
		return &p.verts[len(p.verts)-1]
	}
	return nil
}

// Smooth smoothes this polygon vertex by a radius and discretises the smoothing in facets.
func (v *PolygonControlPoint) Smooth(radius float32, facets int) {
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
func (v *PolygonControlPoint) Arc(radius float32, facets int) {
	if radius != 0 && facets > 0 {
		v.radius = radius
		v.facets = -int32(facets)
	}
}
func (v *PolygonControlPoint) isSmoothed() bool { return v.facets > 0 && v.radius > 0 }
func (v *PolygonControlPoint) isArc() bool      { return v.facets < 0 && v.radius != 0 }

const sqrtHalf = math.Sqrt2 / 2

// Chamfer is a smoothing of a single facet of length `size`.
func (v *PolygonControlPoint) Chamfer(size float32) {
	v.Smooth(size*sqrtHalf, 1)
}

func appendArc2points(dst []Vec, p1, p2 Vec, r float32, facets int32) ([]Vec, error) {
	if facets <= 1 {
		return dst, nil // Nothing to do.
	}
	arcCenter, arcAngle, err := arcCenterFrom2points(p1, p2, r)
	if err != nil {
		return dst, err
	}
	return appendArcWithCenter(dst, p1, arcCenter, arcAngle, facets), nil
}

func appendArcWithCenter(dst []Vec, start, center Vec, arcAngle float32, facets int32) []Vec {
	dtheta := arcAngle / float32(facets)
	T := RotationMat2(dtheta)
	rv := Sub(start, center)
	for i := int32(0); i < facets-1; i++ {
		rv = MulMatVec(T, rv)
		dst = append(dst, Add(center, rv))
	}
	return dst
}

func arcCenterFrom2points(p1, p2 Vec, r float32) (Vec, float32, error) {
	rabs := math.Abs(r)
	V12 := Sub(p2, p1)
	chordCenter := Add(p1, Scale(0.5, V12))
	chordLen := Norm(V12)   // Chord length.
	maxChordLen := 2 * rabs //
	if chordLen-maxChordLen > arcTol {
		return Vec{}, 0, errSmallArcRadius
	}
	// Theta is the opening angle from the center of the arc circle
	// to the two chord points.
	// Due to chord definition theta/2 is the angle formed
	// by the chord and the tangent to the chord point.
	chordThetaDiv2 := math.Asin(chordLen / (2 * rabs))
	diffTo90 := chordThetaDiv2 - math.Pi/2
	if math.Abs(diffTo90) < 1e-6 {
		// Ill conditioned arc. Do a little correction away from the 90 degree mark.
		chordThetaDiv2 += math.Copysign(1e-6, -diffTo90)
	}
	// We now find the arc center. To do this we look at the radius
	// sign which will tell us the orientation of the arc (clockwise vs counterclockwise).
	// Y:-X -> for the simple case V12=Vec{X:1} this results in
	// perp=Vec{Y:-1}. If we place arc center in that direction
	// that will result in a clockwise arc, so negative angle.
	perp := Vec{
		X: math.Copysign(V12.Y, -V12.Y*r),
		Y: math.Copysign(V12.X, V12.X*r),
	}
	// x is distance from arc center to chord center.
	x := 0.5 * chordLen / math.Tan(chordThetaDiv2)
	// Perp is scaled to be of length x.
	// Then, simply add perp for arc center.
	perp = Scale(x/chordLen, perp)
	return Add(chordCenter, perp), math.Copysign(2*chordThetaDiv2, r), nil
}

var (
	errLargeSmoothRadius = errors.New("smoothing radius too large")
	errSmallArcRadius    = errors.New("arc radius too small")
	errBadSmooth         = errors.New("badly conditioned smoothing")
	errCPEqualToPrev     = errors.New("equal to previous control point")
)

func appendSmoothedCorner(dst []Vec, p0, p1, p2 Vec, r float32, facets int32) ([]Vec, error) {

	if facets <= 1 {
		return dst, nil // Chamfer case facets==1.
	}
	// Calculate midpoint between two control points.
	// The arc center of corner will lie in direction of this midpoint from corner point p1.
	V10 := Sub(p0, p1)
	norm10 := Norm(V10)
	V12 := Sub(p2, p1)
	norm12 := Norm(V12)
	if r-norm10 > arcTol || r-norm12 > arcTol {
		return dst, errLargeSmoothRadius
	} else if norm10 == 0 || norm12 == 0 {
		return dst, errBadSmooth
	}
	// Normalize vectors.
	V10 = Scale(1/norm10, V10)
	V12 = Scale(1/norm12, V12)

	// theta is opening angle between two vectors going from smoothed corner p1 to p0 and p2.
	theta := math.Acos(Dot(V10, V12)) // Since normalized no need to divide further.
	if math.Abs(theta) < arcTol {
		return dst, errBadSmooth
	} else if math.Abs(theta-math.Pi) < arcTol {
		return dst, errBadSmooth
	}

	sint, cost := math.Sincos(0.5 * theta)
	d := r / (sint / cost)
	// d is distance from p1 to tangent points on arc.
	start := Add(p1, Scale(d, V10))
	end := Add(p1, Scale(d, V12))
	if !EqualElem(p0, start, arcTol*norm10) {
		dst = append(dst, start) // Cap smooth if p0 point not near radius start.
	}

	// Now find chord midpoint and find arc center using dc, distance from p1 to arc center.
	// chordMidpoint := Scale(0.5, Add(start, end))
	arcCenterDir := Unit(Add(V10, V12))
	arcCenter := Add(p1, Scale(r/sint, arcCenterDir))
	arcAngle := math.Pi - theta
	arcAngle = CopyOrientation(arcAngle, start, p1, end)
	dst = appendArcWithCenter(dst, start, arcCenter, arcAngle, facets)
	if !EqualElem(p2, end, arcTol*norm12) {
		dst = append(dst, end) // Cap smooth if p2 point not near radius end.
	}
	return dst, nil
}
