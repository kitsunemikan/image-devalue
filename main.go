package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/inkyblackness/imgui-go/v4"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ncruces/zenity"
)

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
	defaultWindowTitle  = "Image Devalue"
)

type App struct {
	sourceImage         *ebiten.Image
	sourceImageOp       ebiten.DrawRectShaderOptions
	sourceImageFilename string

	guiFileError string

	mgr *renderer.Manager

	screenW, screenH int

	currentEffect int32
	effects       []*ebiten.Shader
	effectParams  []EffectParams

	Intensity   float32
	TargetValue float32
	Gamma       float32
}

func NewApp() (*App, error) {
	app := &App{
		effects:      make([]*ebiten.Shader, len(effectInfos)),
		effectParams: make([]EffectParams, len(effectInfos)),
	}

	for i := range effectInfos {
		shader, err := ebiten.NewShader(effectInfos[i].ShaderSrc)
		if err != nil {
			return nil, fmt.Errorf("compile '%v' shader: %w", effectInfos[i].Name, err)
		}

		app.effects[i] = shader
		app.effectParams[i] = effectInfos[i].NewParamsFunc()
	}

	app.mgr = renderer.New(nil)

	app.sourceImageOp = ebiten.DrawRectShaderOptions{}
	app.sourceImageOp.Uniforms = make(map[string]any)

	return app, nil
}

func (app *App) loadImage(filename string) error {
	app.guiFileError = ""

	imageFile, err := os.Open(filename)
	if err != nil {
		app.guiFileError = err.Error()
		return fmt.Errorf("open image: %w", err)
	}

	defer imageFile.Close()

	rawImage, imgFmt, err := image.Decode(imageFile)
	if err != nil {
		app.guiFileError = err.Error()
		return fmt.Errorf("decode default image '%s': %w", filename, err)
	}

	app.sourceImageFilename = filename

	app.sourceImage = ebiten.NewImageFromImage(rawImage)
	app.sourceImageOp.Images[0] = app.sourceImage

	app.repositionImage()

	log.Info().
		Str("filepath", filename).
		Str("format", imgFmt).
		Msg("Loaded new image")

	return nil
}

func (app *App) repositionImage() {
	if app.sourceImage == nil {
		return
	}

	imgW, imgH := app.sourceImage.Size()

	ratioX := float64(app.screenW) / float64(imgW)
	ratioY := float64(app.screenH) / float64(imgH)

	fitInsideRatio := 1.0
	if ratioX < 1.0 {
		fitInsideRatio = ratioX
	}

	if ratioY < fitInsideRatio {
		fitInsideRatio = ratioY
	}

	app.sourceImageOp.GeoM.Reset()
	app.sourceImageOp.GeoM.Translate(-float64(imgW), -float64(imgH))
	app.sourceImageOp.GeoM.Scale(fitInsideRatio, fitInsideRatio)
	app.sourceImageOp.GeoM.Translate(float64(app.screenW), float64(app.screenH))
}

func (app *App) Draw(screen *ebiten.Image) {
	if app.sourceImage != nil {
		imgW, imgH := app.sourceImage.Size()
		screen.DrawRectShader(imgW, imgH, app.effects[int(app.currentEffect)], &app.sourceImageOp)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nFPS: %.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))

	app.mgr.Draw(screen)
}

func (app *App) guiOpenImage() {
	log.Info().Msg("Selecting image to open")

	filename, err := zenity.SelectFile(
		zenity.Filename(app.sourceImageFilename),
		zenity.FileFilters{
			{"JPEG (*.jpg, *.jpeg, *.jpe, *.jfif)", []string{"*.jpg", "*.jpeg", "*.jpe", "*.jfif"}, true},
			{"PNG (*.png)", []string{"*.png"}, true},
			{"All files", []string{"*"}, true},
		},
	)

	if err == zenity.ErrCanceled {
		log.Info().Msg("Open image dialog canceled")
		return
	}

	if err != nil {
		log.Err(err).Msg("Select image file to open")
		return
	}

	err = app.loadImage(filename)
	if err != nil {
		log.Err(err).Msg("Open image")
		return
	}
}

func (app *App) exportImage(filename string) error {
	log.Info().Str("file", filename).Msg("Export")

	imageFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}

	defer imageFile.Close()

	imgW, imgH := app.sourceImage.Size()

	postprocessed := ebiten.NewImage(imgW, imgH)

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = make(map[string]any)
	SetUniforms(app.effectParams[int(app.currentEffect)], op)
	op.Images[0] = app.sourceImage

	postprocessed.DrawRectShader(imgW, imgH, app.effects[int(app.currentEffect)], op)

	rawImage := postprocessed.SubImage(postprocessed.Bounds())

	err = png.Encode(imageFile, rawImage)
	if err != nil {
		return fmt.Errorf("encode image: %w", err)
	}

	return nil
}

func (app *App) guiExportImage() {
	app.guiFileError = ""

	if app.sourceImage == nil {
		app.guiFileError = "No image to export yet"
		log.Error().Msg("No image to export yet")
		return
	}

	log.Info().Msg("Selecting export destination")

	preferredDir := filepath.Dir(os.Args[0])
	wd, err := os.Getwd()
	if err != nil {
		preferredDir = wd
	}

	filename, err := zenity.SelectFileSave(
		zenity.Filename(filepath.Join(preferredDir, "devalued_export")),
		zenity.ConfirmOverwrite(),
		zenity.FileFilters{
			{"PNG", []string{"*.png"}, true},
		},
	)

	if err == zenity.ErrCanceled {
		log.Info().Msg("Export image dialog canceled")
		return
	}

	if err != nil {
		log.Err(err).Msg("Select image export destination")
		return
	}

	if filepath.Ext(filename) == "" {
		filename += ".png"
	}

	err = app.exportImage(filename)
	if err != nil {
		app.guiFileError = err.Error()
		log.Err(err).Str("filename", filename).Msg("Export image")
	}
}

func (app *App) Update() error {
	app.mgr.Update(1.0 / 60.0)
	app.mgr.BeginFrame()

	imgui.Bullet()
	imgui.Text("File")

	if app.guiFileError != "" {
		imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{1.0, 0.3, 0.3, 1.0})
		imgui.Text(app.guiFileError)
		imgui.PopStyleColor()
	} else if app.sourceImageFilename == "" {
		imgui.Text("Please, open an image")
	} else {
		imgui.Text(app.sourceImageFilename)
	}

	if imgui.Button("Open") {
		app.guiOpenImage()
	}

	imgui.SameLine()

	if imgui.Button("Export") {
		app.guiExportImage()
	}

	imgui.Separator()
	imgui.Bullet()
	imgui.Text("Transform")

	imgui.Combo("Effect", &app.currentEffect, effectNames)

	sliders := app.effectParams[int(app.currentEffect)].Sliders()
	for i := range sliders {
		imgui.SliderFloat(sliders[i].Title, sliders[i].Target, sliders[i].Min, sliders[i].Max)
	}

	app.mgr.EndFrame()

	SetUniforms(app.effectParams[int(app.currentEffect)], &app.sourceImageOp)
	return nil
}

func (app *App) Layout(outsideW, outsideH int) (int, int) {
	if app.screenW != outsideW || app.screenH != outsideH {
		log.Info().Int("newW", outsideW).Int("newH", outsideH).Msg("Window resized")
		app.screenW, app.screenH = outsideW, outsideH

		app.mgr.SetDisplaySize(float32(outsideW), float32(outsideH))
		app.repositionImage()
	}

	return outsideW, outsideH
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ebiten.SetWindowSize(defaultWindowWidth, defaultWindowHeight)
	ebiten.SetWindowTitle(defaultWindowTitle)
	ebiten.SetWindowResizable(true)

	app, err := NewApp()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Initialize application")
		os.Exit(1)
	}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal().
			Err(err).
			Msg("Run app")
		os.Exit(1)
	}
}
