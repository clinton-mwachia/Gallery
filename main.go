package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	imagePaths     []string
	currentIndex   int
	slideshowOn    bool
	slideshowDur   = 2 * time.Second
	mainWindow     fyne.Window
	imgCanvas      *canvas.Image
	slideshowBtn   *widget.Button
	statusLabel    *widget.Label
	speedSlider    *widget.Slider
	speedLabel     *widget.Label // Displays current slideshow speed
	countdownLabel *widget.Label // Shows countdown before the next image
)

func main() {
	a := app.NewWithID("com.example.gallery")
	mainWindow = a.NewWindow("Gallery Viewer")
	mainWindow.Resize(fyne.NewSize(800, 600))

	// UI Components
	openBtn := widget.NewButton("Open Folder", func() {
		openFolder()
	})

	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		navigateImage(-1)
	})
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		navigateImage(1)
	})

	slideshowBtn = widget.NewButton("Start Slideshow", toggleSlideshow)

	// Image count label
	statusLabel = widget.NewLabel("No images loaded") // Default text

	// Slideshow speed slider
	speedSlider = widget.NewSlider(1, 10) // Range: 1s to 10s
	speedSlider.Value = 2                 // Default speed
	speedSlider.Step = 1
	speedLabel = widget.NewLabel(fmt.Sprintf("Speed: %ds", int(speedSlider.Value)))

	// Countdown label
	countdownLabel = widget.NewLabel("")

	speedSlider.OnChanged = func(value float64) {
		slideshowDur = time.Duration(value) * time.Second
		speedLabel.SetText(fmt.Sprintf("Speed: %ds", int(value)))
	}

	// Image Display
	imgCanvas = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	imgCanvas.FillMode = canvas.ImageFillContain

	// Layout
	controls := container.NewVBox(
		container.NewHBox(prevBtn, statusLabel, nextBtn, slideshowBtn),
		container.NewHBox(widget.NewLabel("Slideshow Speed:"), speedSlider, speedLabel, countdownLabel),
	)
	content := container.NewBorder(openBtn, controls, nil, nil, imgCanvas)

	mainWindow.SetContent(content)
	mainWindow.ShowAndRun()
}

func openFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if uri == nil {
			dialog.ShowInformation("Opening Folder", "Error Opening Folder: "+err.Error(), mainWindow)
			return
		}

		files, _ := os.ReadDir(uri.Path())
		imagePaths = nil

		for _, file := range files {
			ext := filepath.Ext(file.Name())
			if isValidImageExt(strings.ToLower(ext)) {
				imagePaths = append(imagePaths, filepath.Join(uri.Path(), file.Name()))
			}
		}

		if len(imagePaths) == 0 {
			dialog.ShowInformation("No Images Found", "The selected folder contains no images.", mainWindow)
			statusLabel.SetText("No images loaded")
			return
		}
		sort.Strings(imagePaths) // Sort files alphabetically
		if len(imagePaths) > 0 {
			currentIndex = 0
			loadImage()
		}
	}, mainWindow)
}

func isValidImageExt(ext string) bool {
	validExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	}
	return validExts[ext]
}

func loadImage() {
	if len(imagePaths) == 0 || currentIndex < 0 || currentIndex >= len(imagePaths) {
		dialog.ShowInformation("No Images", "Please load images", mainWindow)
	}

	srcFile, err := os.Open(imagePaths[currentIndex])
	if err != nil {
		dialog.ShowInformation("Loading Images", "Error Opening Images: "+err.Error(), mainWindow)
		return
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		dialog.ShowInformation("Decoding Images", "Error Decoding Images: "+err.Error(), mainWindow)
		return
	}

	imgCanvas.Image = img
	imgCanvas.Refresh()

	// Update status label
	statusLabel.SetText(fmt.Sprintf("Image %d of %d", currentIndex+1, len(imagePaths)))
}

func navigateImage(direction int) {
	if len(imagePaths) == 0 {
		dialog.ShowInformation("No Images", "Please load images before starting a slideshow.", mainWindow)
	}
	currentIndex = (currentIndex + direction + len(imagePaths)) % len(imagePaths)
	loadImage()
}

func toggleSlideshow() {
	if len(imagePaths) == 0 {
		dialog.ShowInformation("No Images", "Please load images before starting a slideshow.", mainWindow)
		return
	}

	if slideshowOn {
		slideshowOn = false
		slideshowBtn.SetText("Start Slideshow")
		countdownLabel.SetText("") // Clear countdown when stopping
		return
	}

	slideshowOn = true
	slideshowBtn.SetText("Stop Slideshow")

	go func() {
		for slideshowOn {
			countdown := int(slideshowDur.Seconds())
			for countdown > 0 && slideshowOn {
				countdownLabel.SetText(fmt.Sprintf("Next image in: %d s", countdown))
				time.Sleep(1 * time.Second)
				countdown--
			}

			if slideshowOn { // Only move if slideshow is still on
				navigateImage(1)
			}
		}
	}()
}
