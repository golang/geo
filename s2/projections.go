package s2

import (
	"math"
)

// type Projections int

// const (
// 	S2_LINEAR_PROJECTION Projections = iota
// 	S2_TAN_PROJECTION
// 	S2_QUADRATIC_PROJECTION
// )

// const S2_PROJECTION Projections = S2_QUADRATIC_PROJECTION

var (
	S2_PROJECTION Projections = S2_QUADRATIC_PROJECTION{}
)

type Projections interface {
	MAX_ANGLE_SPAN() Metric
	MIN_WIDTH() Metric
	MAX_WIDTH() Metric
	AVG_WIDTH() Metric
}

type S2_QUADRATIC_PROJECTION struct{}

func (m S2_QUADRATIC_PROJECTION) MAX_ANGLE_SPAN() Metric {
	return Metric{
		0.85244858959960922, // 0.852
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MIN_WIDTH() Metric {
	return Metric{
		math.Sqrt2 / 3, // 0.471
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MAX_WIDTH() Metric {
	return m.MAX_ANGLE_SPAN()
}

func (m S2_QUADRATIC_PROJECTION) AVG_WIDTH() Metric {
	return Metric{
		0.71726183644304969, // 0.717
		1,
	}
}
