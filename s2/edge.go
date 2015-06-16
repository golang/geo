package s2

import (
	"fmt"
)

type Edge struct {
	start Point
	end   Point
}

func NewEdgeFromStartEnd(start, end Point) Edge {
	return Edge{start, end}
}

func (e Edge) Start() Point { return e.start }
func (e Edge) End() Point   { return e.end }

func (e Edge) String() string {
	return fmt.Sprintf("Edge: (%s -> %s)\n   or [%s -> %s]", e.start.DegreesString(), e.end.DegreesString(), e.start.String(), e.end.String())
}

func (e Edge) Equals(other Edge) bool {
	return e.start.ApproxEquals(other.start, EPSILON) && e.end.ApproxEquals(other.end, EPSILON)
}
