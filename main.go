package main

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	// resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/shader"
	//"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/inkyblackness/imgui-go/v4"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
	defaultWindowTitle  = "Image Devalue"

	defaultImagePath = "test_image.jpg"
)

type App struct {
	sourceImage *ebiten.Image
	mgr         *renderer.Manager

	devalueIntensity   float32
	devalueTargetValue float32
}

func NewApp() (*App, error) {
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

	sourceImage := ebiten.NewImageFromImage(defaultImg)

	mgr := renderer.New(nil)
	mgr.SetText("Image devalue")

	return &App{
		sourceImage: sourceImage,
		mgr:         mgr,

		devalueIntensity:   0.0,
		devalueTargetValue: 0.5,
	}, nil
}

func (app *App) Draw(screen *ebiten.Image) {
	screenW, screenH := screen.Size()
	imgW, imgH := app.sourceImage.Size()

	op := ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Scale(float64(screenW)/float64(imgW), float64(screenH)/float64(imgH))

	screen.DrawImage(app.sourceImage, &op)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %.2f\nFPS: %.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))

	app.mgr.Draw(screen)
}

func (app *App) Update() error {
	app.mgr.Update(1.0 / 60.0)
	app.mgr.BeginFrame()

	imgui.Bullet()
	imgui.Text("File")

	imgui.Text("Current: " + defaultImagePath)

	if imgui.Button("Open") {
		log.Info().Msg("Open button pressed")
	}

	imgui.SameLine()

	if imgui.Button("Export") {
		log.Info().Msg("Export requested")
	}

	imgui.Separator()
	imgui.Bullet()
	imgui.Text("Devalue options")

	imgui.SliderFloat("Intensity", &app.devalueIntensity, 0.0, 100.0)
	imgui.SliderFloat("Target value", &app.devalueTargetValue, 0.0, 1.0)

	app.mgr.EndFrame()
	return nil
}

func (app *App) Layout(outsideW, outsideH int) (int, int) {
	app.mgr.SetDisplaySize(float32(outsideW), float32(outsideH))
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
