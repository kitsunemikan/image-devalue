//go:build ignore

//kage:unit pixel

package main

var (
	DevalueIntensity   float
	DevalueTargetValue float
)

// Ported from https://stackoverflow.com/a/6930407
func rgb2hsv(in vec3) vec3 {
	out := vec3(0)

	min := in.r
	if min > in.g {
		min = in.g
	}

	if min > in.b {
		min = in.b
	}

	max := in.r
	if max < in.g {
		max = in.g
	}

	if max < in.b {
		max = in.b
	}

	// Value
	out.z = max

	delta := max - min
	if delta < 0.00001 {
		out.x = 0 // undefined, maybe nan?
		out.y = 0
		return out
	}

	// NOTE: if Max is == 0, this divide would cause a crash
	if max > 0.0 {
		// Saturation
		out.y = delta / max
	} else {
		// if max is 0, then r = g = b = 0
		// s = 0, h is undefined
		out.x = log(-1.0) // its now undefined (NaN)
		out.y = 0.0
		return out
	}

	// > is bogus, just keeps compilor happy
	if in.r >= max {
		// between yellow & magenta
		out.x = (in.g - in.b) / delta
	} else if in.g >= max {
		// between cyan & yellow
		out.x = 2.0 + (in.b-in.r)/delta
	} else {
		// between magenta & cyan
		out.x = 4.0 + (in.r-in.g)/delta
	}

	// Degrees
	out.x *= 60.0

	if out.x < 0.0 {
		out.x += 360.0
	}

	return out
}

func hsv2rgb(in vec3) vec3 {
	out := vec3(0)

	// < is bogus, just shuts up warnings
	if in.y <= 0.0 {
		out.r = in.z
		out.g = in.z
		out.b = in.z
		return out
	}

	hh := in.x

	if hh >= 360.0 {
		hh = 0.0
	}

	hh /= 60.0

	i := floor(hh)
	ff := fract(hh)

	p := in.z * (1.0 - in.y)
	q := in.z * (1.0 - (in.y * ff))
	t := in.z * (1.0 - (in.y * (1.0 - ff)))

	if i == 0 {
		out.r = in.z
		out.g = t
		out.b = p
	} else if i == 1 {
		out.r = q
		out.g = in.z
		out.b = p
	} else if i == 2 {
		out.r = p
		out.g = in.z
		out.b = t
	} else if i == 3 {
		out.r = p
		out.g = q
		out.b = in.z
	} else if i == 4 {
		out.r = t
		out.g = p
		out.b = in.z
	} else {
		out.r = in.z
		out.g = p
		out.b = q
	}

	return out
}

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	clr := imageSrc0At(texCoord)

	hsv := rgb2hsv(clr.rgb)
	hsv.z = hsv.z*(1.0-DevalueIntensity) + DevalueTargetValue*DevalueIntensity

	newColor := hsv2rgb(hsv)
	return vec4(newColor, clr.a)
}
