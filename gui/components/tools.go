package components

import (
	"image/color"
)

func NumTransition(n1, n2, trans float32) float32 {
	dif := n1 - n2
	return n1 - dif*trans
}

func ColorTransition(c1, c2 color.NRGBA, trans float32) color.NRGBA {
	return color.NRGBA{
		uint8(NumTransition(float32(c1.R), float32(c2.R), trans)),
		uint8(NumTransition(float32(c1.G), float32(c2.G), trans)),
		uint8(NumTransition(float32(c1.B), float32(c2.B), trans)),
		uint8(NumTransition(float32(c1.A), float32(c2.A), trans)),
	}
}
