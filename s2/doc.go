/*
Package s2 implements types and functions for working with geometry in S² (spherical geometry).

Its related packages, parallel to this one, are s1 (operates on S¹), r1 (operates on ℝ¹)
and r3 (operates on ℝ³).

This package provides types and functions for the S2 cell hierarchy and coordinate systems.
The S2 cell hierarchy is a hierarchical decomposition of the surface of a unit sphere (S²)
into ``cells''; it is highly efficient, scales from continental size to under 1 cm²
and preserves spatial locality (nearby cells have close IDs).

A presentation that gives an overview of S2 is
https://docs.google.com/presentation/d/1Hl4KapfAENAOf4gv-pSngKwvS_jwNVHRPZTTDzXXn6Q/view.
*/
package s2
