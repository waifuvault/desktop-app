package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strconv"
)

func main() {
	waifuVaultApp := app.New()
	w := waifuVaultApp.NewWindow("WaifuVault")
	w.Resize(fyne.NewSize(500, 100))
	title := canvas.NewText("Open file", color.White)
	file := binding.NewString()
	checkButton := widget.NewButton("Choose file", func() {
		chooseDirectory(w, file)
	})
	windowContent := container.NewVBox(
		createCenteredComponents(title),
		container.NewVBox(checkButton),
	)
	w.SetContent(windowContent)
	w.ShowAndRun()
}

func chooseDirectory(w fyne.Window, h binding.String) {
	currentSize := w.Canvas().Size()
	resetSize := func() {
		w.Resize(fyne.NewSize(currentSize.Width, currentSize.Height))
	}
	w.Resize(fyne.NewSize(925, 532))
	dialog.ShowFileOpen(func(file fyne.URIReadCloser, err error) {
		saveDir := ""
		if err != nil {
			dialog.ShowError(err, w)
			resetSize()
			return
		}

		if file == nil {
			resetSize()
			return
		}

		fmt.Println(file.URI())
		saveDir = file.URI().Path()

		err = h.Set(saveDir)
		if err != nil {
			dialog.ShowError(err, w)
			resetSize()
			return
		}
		resetSize()
	}, w)
}

func onEnter(value string, result binding.String) {
	if value == "" {
		result.Set("value can not be empty")
		return
	}
	s, err := strconv.Atoi(value)
	if err != nil {
		result.Set(value + " is not a valid year")
		return
	}
	result.Set("is " + value + " a leap year? " + strconv.FormatBool(isLeap(s)))
}

func isLeap(year int) bool {
	return year%4 == 0 && year%100 != 0 || year%400 == 0
}
func createCenteredComponents(components ...fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewHBoxLayout(), layout.NewSpacer(), container.NewWithoutLayout(components...), layout.NewSpacer())
}
