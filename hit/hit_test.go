package hit

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewArea(t *testing.T) {
	min, max := Point{X: -20, Y: -15}, Point{X: 45, Y: 55}
	a := NewArea(min, max)

	a.buckets = nil // Too much
	if diff := cmp.Diff(a, &Area{
		min:     min,
		max:     max,
		xStride: 65,
		yStride: 70,
		xLen:    numX,
		yLen:    numY,
	}, cmp.AllowUnexported(Area{})); diff != "" {
		t.Errorf("unexpected area (-got, +want): \n%s", diff)
	}
}

func TestMapToBucket(t *testing.T) {
	tcs := []struct {
		name         string
		min, max     Point
		tp           Point
		wantX, wantY int
	}{
		{
			name:  "edge left",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{0, 50},
			wantX: 0,
			wantY: numY / 2,
		},
		{
			name:  "edge right",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{100, 50},
			wantX: numX - 1,
			wantY: numY / 2,
		},
		{
			name:  "center",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{50, 50},
			wantX: numX / 2,
			wantY: numY / 2,
		},
		{
			name:  "edge top",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{50, 0},
			wantX: numX / 2,
			wantY: 0,
		},
		{
			name:  "edge bottom",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{50, 100},
			wantX: numX / 2,
			wantY: numY - 1,
		},
		{
			name:  "low oob",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{-50, -50},
			wantX: 0,
			wantY: 0,
		},
		{
			name:  "high oob",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{500, 500},
			wantX: numX - 1,
			wantY: numY - 1,
		},
		{
			name:  "one oob",
			min:   Point{0, 0},
			max:   Point{100, 100},
			tp:    Point{50, 500},
			wantX: numX / 2,
			wantY: numY - 1,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			a := NewArea(tc.min, tc.max)
			if gotX, gotY := a.mapToBucket(tc.tp); gotX != tc.wantX || gotY != tc.wantY {
				t.Errorf("mapToBucket(%v) = (%d,%d), want (%d,%d)", tc.tp, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestAreaAdd(t *testing.T) {
	tcs := []struct {
		name        string
		min, max    Point
		wantBuckets [][]int
	}{
		{
			name:        "single point low",
			min:         Point{0, 0},
			max:         Point{0, 0},
			wantBuckets: [][]int{[]int{0, 0}},
		},
		{
			name:        "single point mid",
			min:         Point{50, 50},
			max:         Point{50, 50},
			wantBuckets: [][]int{[]int{numX / 2, numY / 2}},
		},
		{
			name:        "single point high",
			min:         Point{100, 100},
			max:         Point{100, 100},
			wantBuckets: [][]int{[]int{numX - 1, numY - 1}},
		},
		{
			name:        "single point oob",
			min:         Point{10000, 10000},
			max:         Point{10000, 10000},
			wantBuckets: [][]int{[]int{numX - 1, numY - 1}},
		},
		{
			name:        "line x",
			min:         Point{0, 0},
			max:         Point{8, 0},
			wantBuckets: [][]int{{0, 0}, {1, 0}},
		},
		{
			name:        "line y",
			min:         Point{0, 0},
			max:         Point{0, 18},
			wantBuckets: [][]int{{0, 0}, {0, 1}},
		},
		{
			name:        "line x long",
			min:         Point{0, 0},
			max:         Point{38, 0},
			wantBuckets: [][]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0}, {6, 0}},
		},
		{
			name:        "line y long",
			min:         Point{0, 0},
			max:         Point{0, 55},
			wantBuckets: [][]int{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}},
		},
		{
			name:        "block small",
			min:         Point{0, 0},
			max:         Point{10, 13},
			wantBuckets: [][]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
		},
		{
			name:        "block med x",
			min:         Point{0, 0},
			max:         Point{17, 13},
			wantBuckets: [][]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 0}, {2, 1}},
		},
		{
			name:        "block med y",
			min:         Point{0, 0},
			max:         Point{10, 25},
			wantBuckets: [][]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}},
		},
		{
			name:        "block tiny ",
			min:         Point{8, 13},
			max:         Point{11, 14},
			wantBuckets: [][]int{{1, 1}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			a := NewArea(Point{}, Point{100, 100})
			a.Add(tc.min, tc.max, nil)

			hadBuckets := make([][]int, 0)
			for x := 0; x < a.xLen; x++ {
				for y := 0; y < a.yLen; y++ {
					if len(a.buckets[x][y].objs) > 0 {
						hadBuckets = append(hadBuckets, []int{x, y})
					}
				}
			}

			if diff := cmp.Diff(tc.wantBuckets, hadBuckets); diff != "" {
				t.Errorf("Different buckets populated than expected (+got, -want): %s\n", diff)
			}
		})
	}
}

type dummyObj struct {
	min, max Point
}

func (d dummyObj) HitTest(p Point) bool {
	return p.X >= d.min.X && p.X <= d.max.X && p.Y >= d.min.Y && p.Y <= d.max.Y
}

func TestAreaHitTest(t *testing.T) {
	a := NewArea(Point{}, Point{400, 400})

	d := dummyObj{Point{35, 55}, Point{121, 185}}
	a.Add(d.min, d.max, d)

	if a.Test(Point{0, 0}) != nil {
		t.Error("Test(0,0) did not return nil")
	}
	if a.Test(Point{100, 0}) != nil {
		t.Error("Test(100,0) did not return nil")
	}
	for x := d.min.X; x < d.max.X; x++ {
		for y := d.min.Y; y < d.max.Y; y++ {
			if got := a.Test(Point{x, y}); got != d {
				t.Errorf("Test(%v,%v) did not return inserted object, got %v", x, y, got)
			}
		}
	}

	d2 := dummyObj{Point{45, 65}, Point{111, 175}}
	a.Add(d2.min, d2.max, d2)
	for x := d2.min.X; x < d2.max.X; x++ {
		for y := d2.min.Y; y < d2.max.Y; y++ {
			if got := a.Test(Point{x, y}); got != d2 {
				t.Errorf("Test(%v,%v) did not return inserted object, got %v", x, y, got)
			}
		}
	}
	if a.Test(Point{100, 0}) != nil {
		t.Error("Test(100,0) did not return nil")
	}
	if got := a.Test(Point{38, 55}); got != d {
		t.Errorf("Test(38,55) = %v, want %v", got, d)
	}
}
