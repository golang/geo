package s2

import (
	"math"
)

var (
	// Note that other parts of this GO port have been hardcoded to QUADRATIC projection

	// Uncomment the desirect projection type
	// S2_PROJECTION Projections = S2_LINEAR_PROJECTION{}
	// S2_PROJECTION Projections = S2_TAN_PROJECTION{}
	S2_PROJECTION Projections = S2_QUADRATIC_PROJECTION{}
)

type Projections interface {
	MIN_AREA() Metric
	MAX_AREA() Metric
	AVG_AREA() Metric

	MIN_ANGLE_SPAN() Metric
	MAX_ANGLE_SPAN() Metric
	AVG_ANGLE_SPAN() Metric

	MIN_WIDTH() Metric
	MAX_WIDTH() Metric
	AVG_WIDTH() Metric

	MIN_EDGE() Metric
	MAX_EDGE() Metric
	AVG_EDGE() Metric

	MIN_DIAG() Metric
	MAX_DIAG() Metric
	AVG_DIAG() Metric

	MAX_EDGE_ASPECT() float64
	MAX_DIAG_ASPECT() float64
}

type S2_PROJECTION_COMMON struct{}

func (m S2_PROJECTION_COMMON) AVG_AREA() Metric {
	return Metric{
		math.Pi / 6, // 0.524
		2,
	}
}

func (m S2_PROJECTION_COMMON) AVG_ANGLE_SPAN() Metric {
	return Metric{
		math.Pi / 4, // 0.785
		1,
	}
}

func (m S2_PROJECTION_COMMON) MAX_DIAG_ASPECT() float64 {
	return math.Sqrt(3) // 1.732
}

type S2_QUADRATIC_PROJECTION struct{ S2_PROJECTION_COMMON }

func (m S2_QUADRATIC_PROJECTION) MIN_AREA() Metric {
	return Metric{
		2 * math.Sqrt2 / 9, // 0.314
		2,
	}
}

func (m S2_QUADRATIC_PROJECTION) MAX_AREA() Metric {
	return Metric{
		0.65894981424079037, // 0.659
		2,
	}
}

func (m S2_QUADRATIC_PROJECTION) MIN_ANGLE_SPAN() Metric {
	return Metric{
		2.0 / 3.0, // 0.667
		1,
	}
}

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

func (m S2_QUADRATIC_PROJECTION) MIN_EDGE() Metric {
	return Metric{
		math.Sqrt2 / 3, // 0.471
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MAX_EDGE() Metric {
	return m.MAX_ANGLE_SPAN()
}

func (m S2_QUADRATIC_PROJECTION) AVG_EDGE() Metric {
	return Metric{
		0.72960687319305303, // 0.730
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MIN_DIAG() Metric {
	return Metric{
		4 * math.Sqrt2 / 9, // // 0.629
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MAX_DIAG() Metric {
	return Metric{
		1.2193272972170106, // 1.219
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) AVG_DIAG() Metric {
	return Metric{
		1.03021136949923584, // 1.030
		1,
	}
}

func (m S2_QUADRATIC_PROJECTION) MAX_EDGE_ASPECT() float64 {
	return 1.44261527445268292 // 1.443
}

// TODO: Add S2_LINEAR_PROJECTION, S2_TAN_PROJECTION
// type S2_LINEAR_PROJECTION struct{ S2_PROJECTION_COMMON }
// type S2_TAN_PROJECTION struct{ S2_PROJECTION_COMMON }
