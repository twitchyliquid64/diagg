package hit

const (
	numX = 32
	numY = 24
)

type Point struct {
	X, Y float64
}

func (p Point) Sub(in Point) Point {
	return Point{X: p.X - in.X, Y: p.Y - in.Y}
}

type TestableObj interface {
	HitTest(Point) bool
}

type object struct {
	min, max Point
	obj      TestableObj
}

type bucket struct {
	objs []object
}

func (b *bucket) add(min, max Point, obj TestableObj) {
	if b.objs == nil {
		b.objs = make([]object, 0, 6)
	}
	b.objs = append(b.objs, object{min, max, obj})
}

func (b *bucket) test(p Point) TestableObj {
	for i := len(b.objs) - 1; i >= 0; i-- {
		o := b.objs[i]
		if o.min.X <= p.X && o.min.Y <= p.Y && o.max.X >= p.X && o.max.Y >= p.Y {
			if o.obj.HitTest(p) {
				return o.obj
			}
		}
	}
	return nil
}

// Area encapsulates a hit testing region.
type Area struct {
	min, max         Point
	xStride, yStride float64

	xLen, yLen int
	// buckets maps an object into X/Y buckets.
	buckets [][]bucket
}

// Add inserts the given object into the hit testing area.
func (a *Area) Add(min, max Point, obj TestableObj) {
	minX, minY := a.mapToBucket(min)
	maxX, maxY := a.mapToBucket(max)
	xStride, yStride := maxX-minX, maxY-minY

	switch {
	case xStride > 0 && yStride > 0: // Covers buckets in both X and Y dimensions.
		for xStride >= 0 {
			for y := yStride; y >= 0; y-- {
				a.buckets[minX+xStride][minY+y].add(min, max, obj)
			}
			xStride--
		}

	case xStride > 0: // Covers buckets only in X dimension.
		for xStride >= 0 {
			a.buckets[minX+xStride][minY].add(min, max, obj)
			xStride--
		}

	case yStride > 0: // Covers buckets only in Y dimension.
		for yStride >= 0 {
			a.buckets[minX][minY+yStride].add(min, max, obj)
			yStride--
		}

	default:
		a.buckets[minX][minY].add(min, max, obj)
	}
}

// Test computes the topmost object which intersects the given point.
func (a *Area) Test(p Point) TestableObj {
	x, y := a.mapToBucket(p)
	return a.buckets[x][y].test(p)
}

func (a *Area) mapToBucket(p Point) (xIdx, yIdx int) {
	p = p.Sub(a.min)
	x, y := int(p.X*float64(a.xLen)/a.xStride), int(p.Y*float64(a.yLen)/a.yStride)

	// Clamp to bounds.
	if x >= a.xLen {
		x = a.xLen - 1
	}
	if y >= a.yLen {
		y = a.yLen - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

func NewArea(min, max Point) *Area {
	a := &Area{
		min:     min,
		max:     max,
		xStride: max.X - min.X,
		yStride: max.Y - min.Y,
		xLen:    numX,
		yLen:    numY,
		buckets: make([][]bucket, numX),
	}
	for i := range a.buckets {
		a.buckets[i] = make([]bucket, numY)
	}

	a.buckets[0][0].objs = make([]object, 0, 4)
	a.buckets[a.xLen-1][0].objs = make([]object, 0, 4)
	a.buckets[0][a.yLen-1].objs = make([]object, 0, 4)
	a.buckets[a.xLen-1][a.yLen-1].objs = make([]object, 0, 4)

	return a
}
