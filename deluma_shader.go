//go:build ignore

//kage:unit pixel

package main

var (
	Intensity  float
	TargetLuma float
	Gamma      float
)

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	clr := imageSrc0At(texCoord)

	colorSum := 0.2126*pow(clr.r, Gamma) + 0.7152*pow(clr.g, Gamma) + 0.0722*pow(clr.b, Gamma)
	r := 1 - pow(TargetLuma/colorSum, 1.0/Gamma)

	clr.rgb -= r * Intensity * clr.rgb
	return clr
}
