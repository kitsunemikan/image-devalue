package main

import (
	"fmt"
	"image"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	//"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	// resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/shader"
	//"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
	defaultWindowTitle  = "Image Devalue"

	defaultImagePath = "test_image.jpg"
)

type App struct {
	sourceImage *ebiten.Image
}

func NewApp() (*App, error) {
	defaultImgFile, err := os.Open(defaultImagePath)
	if err != nil {
		return nil, fmt.Errorf("open default image: %w", err)
	}

	defer defaultImgFile.Close()

	defaultImg, _, err := image.Decode(defaultImgFile)
	if err != nil {
		return nil, fmt.Errorf("decode default image '%s': %w", defaultImagePath, err)
	}

	sourceImage := ebiten.NewImageFromImage(defaultImg)
	return &App{
		sourceImage: sourceImage,
	}, nil
}

func (app *App) Draw(screen *ebiten.Image) {
	screenW, screenH := screen.Size()
	imgW, imgH := app.sourceImage.Size()

	op := ebiten.DrawImageOptions{}
	op.GeoM.Reset()
	op.GeoM.Scale(float64(screenW)/float64(imgW), float64(screenH)/float64(imgH))

	screen.DrawImage(app.sourceImage, &op)
}

func (app *App) Update() error {
	return nil
}

func (app *App) Layout(outsideW, outsideH int) (int, int) {
	return outsideW, outsideH
}

func main() {
	ebiten.SetWindowSize(defaultWindowWidth, defaultWindowHeight)
	ebiten.SetWindowTitle(defaultWindowTitle)
	ebiten.SetWindowResizable(true)

	app, err := NewApp()
	if err != nil {
		fmt.Printf("error: init application: %v", err)
		os.Exit(1)
	}

	if err := ebiten.RunGame(app); err != nil {
		fmt.Printf("error: run app: %v", err)
		os.Exit(1)
	}
}
