package main

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed devalue_shader.go
	devalueEffectSrc []byte

	//go:embed deluma_shader.go
	delumaEffectSrc []byte
)

type Float32Interval struct {
	Min, Max float32
}

type EffectInfo struct {
	Name          string
	ShaderSrc     []byte
	NewParamsFunc func() EffectParams
}

var effectNames = []string{
	"Devalue in HSV space",
	"Deluma in YUV space",
}

var effectInfos = []EffectInfo{
	{
		Name:          effectNames[0],
		ShaderSrc:     devalueEffectSrc,
		NewParamsFunc: NewDevalueEffectParams,
	},
	{
		Name:          effectNames[1],
		ShaderSrc:     delumaEffectSrc,
		NewParamsFunc: NewDelumaEffectParams,
	},
}

type SliderDescriptor struct {
	Target      *float32
	Title       string
	UniformName string
	Min, Max    float32
}

type EffectParams interface {
	Sliders() []SliderDescriptor
}

func SetUniforms(ep EffectParams, op *ebiten.DrawRectShaderOptions) {
	sliders := ep.Sliders()
	for i := range sliders {
		op.Uniforms[sliders[i].UniformName] = *sliders[i].Target
	}
}

type DevalueEffectParams struct {
	Intensity, TargetValue float32

	sliders []SliderDescriptor
}

func NewDevalueEffectParams() EffectParams {
	p := &DevalueEffectParams{
		Intensity:   1.0,
		TargetValue: 0.5,
	}

	p.sliders = []SliderDescriptor{
		{
			Target:      &p.Intensity,
			Title:       "Devalue intensity",
			UniformName: "Intensity",
			Min:         0.0,
			Max:         1.0,
		},
		{
			Target:      &p.TargetValue,
			Title:       "Target value",
			UniformName: "TargetValue",
			Min:         0.0,
			Max:         1.0,
		},
	}

	return p
}

func (ep *DevalueEffectParams) Sliders() []SliderDescriptor {
	return ep.sliders
}

type DelumaEffectParams struct {
	Intensity, TargetLuma, Gamma float32

	sliders []SliderDescriptor
}

func NewDelumaEffectParams() EffectParams {
	p := &DelumaEffectParams{
		Intensity:  1.0,
		TargetLuma: 0.5,
		Gamma:      1.0,
	}

	p.sliders = []SliderDescriptor{
		{
			Target:      &p.Intensity,
			Title:       "Deluma intensity",
			UniformName: "Intensity",
			Min:         0.0,
			Max:         1.0,
		},
		{
			Target:      &p.TargetLuma,
			Title:       "Target luma",
			UniformName: "TargetLuma",
			Min:         0.0,
			Max:         1.0,
		},
		{
			Target:      &p.Gamma,
			Title:       "Gamma correction",
			UniformName: "Gamma",
			Min:         0.01,
			Max:         4.0,
		},
	}

	return p
}

func (ep *DelumaEffectParams) Sliders() []SliderDescriptor {
	return ep.sliders
}
