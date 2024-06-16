package main

import (
	"context"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	waifuVault "github.com/waifuvault/waifuVault-go-api/pkg"
	waifuMod "github.com/waifuvault/waifuVault-go-api/pkg/mod"
	"image/color"
	"net/http"
	"os"
)

func main() {
	api := waifuVault.NewWaifuvaltApi(http.Client{})

	waifuVaultApp := app.New()
	w := waifuVaultApp.NewWindow("WaifuVault")
	w.Resize(fyne.NewSize(500, 100))
	title := canvas.NewText("Open file", color.White)
	file := binding.NewString()

	checkButton := widget.NewButton("Choose or drag and drop file", func() {
		chooseDirectory(w, file)
	})

	w.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		if len(uris) != 1 {
			dialog.ShowError(errors.New("only one file is allowed"), w)
			return
		}
		path := uris[0].Path()
		file.Set(path)
	})

	resultText := binding.NewString()

	copyButton := widget.NewButtonWithIcon("copy Url", theme.ContentCopyIcon(), func() {
		if content, err := resultText.Get(); err == nil {
			w.Clipboard().SetContent(content)
		}
	})
	copyButton.Disable()

	var uploadButton *widget.Button
	uploadButton = widget.NewButton("Upload file", func() {
		fileToUpload, _ := file.Get()
		if fileToUpload == "" {
			dialog.ShowError(errors.New("no file selected"), w)
			return
		}
		uploadButton.Disable()
		uploadFile(fileToUpload, w, api, func(result *waifuMod.WaifuResponse[string], err error) {
			if err != nil {
				uploadButton.Enable()
				dialog.ShowError(err, w)
				return
			}

			/*url := result.URL
			token := result.Token*/

			copyButton.Enable()
			uploadButton.Enable()
		})
	})
	uploadButton.Disable()

	file.AddListener(binding.NewDataListener(func() {
		str, _ := file.Get()
		if str != "" {
			uploadButton.Enable()
		}
	}))

	selectedFileLabel := widget.NewLabelWithData(file)
	windowContent := container.NewVBox(
		createCenteredComponents(title),
		container.NewVBox(checkButton),
		createCenteredComponents(selectedFileLabel),
		container.NewVBox(uploadButton),
	)

	txtWid := widget.NewLabelWithData(resultText)

	bottomBox := container.NewVBox(
		layout.NewSpacer(),
		copyButton,
	)

	content := container.NewBorder(windowContent, bottomBox, nil, nil, txtWid)

	w.SetContent(content)
	w.ShowAndRun()
}

func uploadFile(
	path string,
	w fyne.Window,
	api waifuMod.Waifuvalt,
	callBack func(*waifuMod.WaifuResponse[string], error),
) {

	fileStruc, err := os.Open(path)
	if err != nil {
		callBack(nil, err)
		return
	}
	resp, err := api.UploadFile(context.TODO(), waifuMod.WaifuvaultPutOpts{
		File: fileStruc,
	})
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	callBack(resp, nil)
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

func createCenteredComponents(components ...fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewHBoxLayout(), layout.NewSpacer(), container.NewWithoutLayout(components...), layout.NewSpacer())
}
