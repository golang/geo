package s2

import "testing"

func BenchmarkCalculateDivisionsByEdge(b *testing.B) {
	type fields struct {
		Shape   Shape
		ChainID int
	}
	type args struct {
		divisionsCount int
	}

	points := make([]Point, 100)

	for i := 0; i < len(points); i++ {
		points[i] = randomPoint()
	}

}
