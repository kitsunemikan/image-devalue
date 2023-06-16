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

	defaultImagePath = "test_image.jpg"
)

//go:embed devalue_shader.go
var devalueShaderSrc []byte

type App struct {
	sourceImage          *ebiten.Image
	currentImageFilename string

	mgr *renderer.Manager

	winW, winH int

	devalueShader      *ebiten.Shader
	devalueIntensity   float32
	devalueTargetValue float32
}

func NewApp() (*App, error) {
	shader, err := ebiten.NewShader(devalueShaderSrc)
	if err != nil {
		return nil, fmt.Errorf("compile devalue shader: %w", err)
	}

	defaultImgFile, err := os.Open(defaultImagePath)
	if err != nil {
		return nil, fmt.Errorf("open default image: %w", err)
	}

	defer defaultImgFile.Close()

	defaultImg, imgFmt, err := image.Decode(defaultImgFile)
	if err != nil {
		return nil, fmt.Errorf("decode default image '%s': %w", defaultImagePath, err)
	}

	log.Info().
		Str("filepath", defaultImagePath).
		Str("format", imgFmt).
		Msg("Loaded default image")

	currentImageFilename := defaultImagePath
	sourceImage := ebiten.NewImageFromImage(defaultImg)

	mgr := renderer.New(nil)
	mgr.SetText("Image devalue")

	return &App{
		sourceImage:          sourceImage,
		currentImageFilename: currentImageFilename,

		mgr: mgr,

		devalueShader:      shader,
		devalueIntensity:   0.0,
		devalueTargetValue: 0.5,
	}, nil
}

func (app *App) Draw(screen *ebiten.Image) {
	screenW, screenH := screen.Size()
	imgW, imgH := app.sourceImage.Size()

	ratioX := float64(screenW) / float64(imgW)
	ratioY := float64(screenH) / float64(imgH)

	fitInsideRatio := 1.0
	if ratioX < 1.0 {
		fitInsideRatio = ratioX
	}

	if ratioY < fitInsideRatio {
		fitInsideRatio = ratioY
	}

	op := ebiten.DrawRectShaderOptions{}
	op.GeoM.Reset()
	op.GeoM.Translate(-float64(imgW), -float64(imgH))
	op.GeoM.Scale(fitInsideRatio, fitInsideRatio)
	op.GeoM.Translate(float64(screenW), float64(screenH))

	op.Uniforms = make(map[string]any)
	op.Uniforms["DevalueIntensity"] = app.devalueIntensity
	op.Uniforms["DevalueTargetValue"] = app.devalueTargetValue

	op.Images[0] = app.sourceImage

	screen.DrawRectShader(imgW, imgH, app.devalueShader, &op)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nFPS: %.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))

	app.mgr.Draw(screen)
}

func (app *App) Update() error {
	app.mgr.Update(1.0 / 60.0)
	app.mgr.BeginFrame()

	imgui.Bullet()
	imgui.Text("File")

	imgui.Text("Current: " + app.currentImageFilename)

	if imgui.Button("Open") {
		log.Info().Msg("Selecting image to open")

		filename, err := zenity.SelectFile(
			zenity.Filename(app.currentImageFilename),
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
			app.currentImageFilename = filename
		}
	}

	imgui.SameLine()

	if imgui.Button("Export") {
		log.Info().Msg("Selecting export destination")

		filename, err := zenity.SelectFileSave(
			zenity.Filename(app.currentImageFilename),
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
	if app.winW != outsideW || app.winH != outsideH {
		log.Info().Int("newW", outsideW).Int("newH", outsideH).Msg("Window resized")
		app.winW, app.winH = outsideW, outsideH

		app.mgr.SetDisplaySize(float32(outsideW), float32(outsideH))
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
