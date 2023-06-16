//go:build ignore

//kage:unit pixel

package main

var (
	DevalueIntensity   float
	DevalueTargetValue float
)

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	clr := imageSrc0At(texCoord)

	clr.rgb = clr.rgb*vec3(1.0-DevalueIntensity) + vec3(DevalueTargetValue)*vec3(DevalueIntensity)
	return clr
}
