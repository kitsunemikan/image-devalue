package main

import (
	_ "embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	// resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/shader"
	//"github.com/hajimehoshi/ebiten/v2/inpututil"

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

//go:embed devalue_shader.go
var devalueShaderSrc []byte

type App struct {
	sourceImage         *ebiten.Image
	sourceImageOp       ebiten.DrawRectShaderOptions
	sourceImageFilename string

	mgr *renderer.Manager

	screenW, screenH int

	devalueShader      *ebiten.Shader
	devalueIntensity   float32
	devalueTargetValue float32
}

func NewApp() (*App, error) {
	shader, err := ebiten.NewShader(devalueShaderSrc)
	if err != nil {
		return nil, fmt.Errorf("compile devalue shader: %w", err)
	}

	mgr := renderer.New(nil)
	mgr.SetText("Image devalue")

	op := ebiten.DrawRectShaderOptions{}
	op.Uniforms = make(map[string]any)

	return &App{
		sourceImage:         nil,
		sourceImageOp:       op,
		sourceImageFilename: "",

		mgr: mgr,

		devalueShader:      shader,
		devalueIntensity:   0.0,
		devalueTargetValue: 0.5,
	}, nil
}

func (app *App) loadImage(filename string) error {
	imageFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open image: %w", err)
	}

	defer imageFile.Close()

	rawImage, imgFmt, err := image.Decode(imageFile)
	if err != nil {
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
		app.sourceImageOp.Uniforms["DevalueIntensity"] = app.devalueIntensity
		app.sourceImageOp.Uniforms["DevalueTargetValue"] = app.devalueTargetValue

		screen.DrawRectShader(imgW, imgH, app.devalueShader, &app.sourceImageOp)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nFPS: %.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))

	app.mgr.Draw(screen)
}

func (app *App) guiOpenImage() {
	log.Info().Msg("Selecting image to open")

	filename, err := zenity.SelectFile(
		zenity.Filename(app.sourceImageFilename),
		zenity.FileFilters{
			{"JPEG", []string{"*.jpg", "*.jpeg", "*.jpe", "*.jfif"}, true},
			{"PNG", []string{"*.png"}, true},
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

func (app *App) Update() error {
	app.mgr.Update(1.0 / 60.0)
	app.mgr.BeginFrame()

	imgui.Bullet()
	imgui.Text("File")

	imgui.Text(app.sourceImageFilename)

	if imgui.Button("Open") {
		app.guiOpenImage()
	}

	imgui.SameLine()

	if imgui.Button("Export") {
		log.Info().Msg("Selecting export destination")

		filename, err := zenity.SelectFileSave(
			zenity.Filename(app.sourceImageFilename),
			zenity.ConfirmOverwrite(),
			zenity.FileFilters{
				{"JPEG", []string{"*.jpg", "*.jpeg", "*.jpe", "*.jfif"}, true},
				{"PNG", []string{"*.png"}, true},
			},
		)

		if err == zenity.ErrCanceled {
			log.Info().Msg("Open image dialog canceled")
		} else if err != nil {
			log.Err(err).Msg("Select image file to open")
		} else {
			log.Info().Str("file", filename).Msg("Export")
		}
	}

	imgui.Separator()
	imgui.Bullet()
	imgui.Text("Devalue options")

	imgui.SliderFloat("Intensity", &app.devalueIntensity, 0.0, 1.0)
	imgui.SliderFloat("Target value", &app.devalueTargetValue, 0.0, 1.0)

	app.mgr.EndFrame()
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
