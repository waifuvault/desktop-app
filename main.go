package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	waifuVault "github.com/waifuvault/waifuVault-go-api/pkg"
	waifuMod "github.com/waifuvault/waifuVault-go-api/pkg/mod"
)

type AppState struct {
	api            waifuMod.Waifuvalt
	window         fyne.Window
	filePath       binding.String
	fileName       binding.String
	fileSize       binding.String
	urlBind        binding.String
	tokenBind      binding.String
	isUploading    binding.Bool
	uploadProgress binding.Float
	uploadComplete binding.Bool
	statusMessage  binding.String
}

func main() {
	api := waifuVault.NewWaifuvaltApi(http.Client{})

	waifuVaultApp := app.NewWithID("moe.waifuvault.desktop")
	w := waifuVaultApp.NewWindow("WaifuVault - File Sharing Made Easy")
	w.Resize(fyne.NewSize(650, 450))
	w.CenterOnScreen()
	w.SetFixedSize(false)

	state := &AppState{
		api:            api,
		window:         w,
		filePath:       binding.NewString(),
		fileName:       binding.NewString(),
		fileSize:       binding.NewString(),
		urlBind:        binding.NewString(),
		tokenBind:      binding.NewString(),
		isUploading:    binding.NewBool(),
		uploadProgress: binding.NewFloat(),
		uploadComplete: binding.NewBool(),
		statusMessage:  binding.NewString(),
	}

	state.statusMessage.Set("Ready to upload")

	content := buildUI(state)
	w.SetContent(content)

	// Handle drag and drop
	w.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		if len(uris) != 1 {
			showNotification(state, "Please drop only one file at a time", true)
			return
		}
		handleFileSelection(state, uris[0].Path())
	})

	w.ShowAndRun()
}

func buildUI(state *AppState) fyne.CanvasObject {
	// Header
	header := createHeader()

	// File drop zone
	dropZone := createDropZone(state)

	// File info card
	fileInfoCard := createFileInfoCard(state)

	// Progress indicator
	progressCard := createProgressCard(state)

	// Upload button
	uploadBtn := createUploadButton(state)

	// Results card
	resultsCard := createResultsCard(state)

	// Status bar
	statusBar := createStatusBar(state)

	mainContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		dropZone,
		fileInfoCard,
		progressCard,
		uploadBtn,
		resultsCard,
	)

	return container.NewBorder(nil, statusBar, nil, nil,
		container.NewPadded(mainContent))
}

func createHeader() fyne.CanvasObject {
	title := widget.NewLabel("WaifuVault File Sharing")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Secure, fast and easy file sharing")
	subtitle.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		title,
		subtitle,
	)
}

func createDropZone(state *AppState) fyne.CanvasObject {
	uploadIcon := widget.NewIcon(theme.UploadIcon())

	dropText := widget.NewLabel("Drag and drop your file here")
	dropText.Alignment = fyne.TextAlignCenter

	selectBtn := widget.NewButton("Select File", func() {
		chooseFile(state)
	})

	dropContent := container.NewVBox(
		container.NewHBox(
			layout.NewSpacer(),
			uploadIcon,
			dropText,
			layout.NewSpacer(),
		),
		container.NewCenter(selectBtn),
	)

	return widget.NewCard("", "", dropContent)
}

func createFileInfoCard(state *AppState) fyne.CanvasObject {
	fileNameLabel := widget.NewLabelWithData(state.fileName)
	fileNameLabel.TextStyle = fyne.TextStyle{Bold: true}
	fileNameLabel.Wrapping = fyne.TextWrapBreak

	fileSizeLabel := widget.NewLabelWithData(state.fileSize)
	fileSizeLabel.TextStyle = fyne.TextStyle{Italic: true}

	fileIcon := widget.NewIcon(theme.FileIcon())

	infoContent := container.NewBorder(
		nil, nil,
		fileIcon,
		nil,
		container.NewVBox(fileNameLabel, fileSizeLabel),
	)

	card := widget.NewCard("", "", infoContent)

	cardContainer := container.NewVBox(card)
	cardContainer.Hide()

	state.fileName.AddListener(binding.NewDataListener(func() {
		name, _ := state.fileName.Get()
		if name != "" {
			cardContainer.Show()
		} else {
			cardContainer.Hide()
		}
	}))

	return cardContainer
}

func createProgressCard(state *AppState) fyne.CanvasObject {
	progressBar := widget.NewProgressBarWithData(state.uploadProgress)

	statusLabel := widget.NewLabelWithData(state.statusMessage)
	statusLabel.Alignment = fyne.TextAlignCenter

	progressContent := container.NewVBox(
		statusLabel,
		progressBar,
	)

	progressContainer := container.NewVBox(progressContent)
	progressContainer.Hide()

	state.isUploading.AddListener(binding.NewDataListener(func() {
		uploading, _ := state.isUploading.Get()
		if uploading {
			progressContainer.Show()
		} else {
			progressContainer.Hide()
		}
	}))

	return progressContainer
}

func createUploadButton(state *AppState) fyne.CanvasObject {
	uploadBtn := widget.NewButtonWithIcon("Upload File", theme.UploadIcon(), func() {
		performUpload(state)
	})
	uploadBtn.Importance = widget.HighImportance
	uploadBtn.Disable()

	state.filePath.AddListener(binding.NewDataListener(func() {
		path, _ := state.filePath.Get()
		uploading, _ := state.isUploading.Get()
		if path != "" && !uploading {
			uploadBtn.Enable()
		} else {
			uploadBtn.Disable()
		}
	}))

	state.isUploading.AddListener(binding.NewDataListener(func() {
		uploading, _ := state.isUploading.Get()
		if uploading {
			uploadBtn.Disable()
		} else {
			path, _ := state.filePath.Get()
			if path != "" {
				uploadBtn.Enable()
			}
		}
	}))

	return container.NewCenter(uploadBtn)
}

func createResultsCard(state *AppState) fyne.CanvasObject {
	urlEntry := widget.NewEntryWithData(state.urlBind)
	urlEntry.Disable()

	tokenEntry := widget.NewEntryWithData(state.tokenBind)
	tokenEntry.Disable()

	copyUrlBtn := widget.NewButtonWithIcon("Copy URL", theme.ContentCopyIcon(), func() {
		url, _ := state.urlBind.Get()
		state.window.Clipboard().SetContent(url)
		showNotification(state, "URL copied to clipboard!", false)
	})

	copyTokenBtn := widget.NewButtonWithIcon("Copy Token", theme.ContentCopyIcon(), func() {
		token, _ := state.tokenBind.Get()
		state.window.Clipboard().SetContent(token)
		showNotification(state, "Token copied to clipboard!", false)
	})

	copyBothBtn := widget.NewButtonWithIcon("Copy Both", theme.ContentCopyIcon(), func() {
		url, _ := state.urlBind.Get()
		token, _ := state.tokenBind.Get()
		both := fmt.Sprintf("URL: %s\nToken: %s", url, token)
		state.window.Clipboard().SetContent(both)
		showNotification(state, "URL and token copied to clipboard!", false)
	})

	urlRow := container.NewBorder(nil, nil, widget.NewLabel("URL:"), copyUrlBtn, urlEntry)
	tokenRow := container.NewBorder(nil, nil, widget.NewLabel("Token:"), copyTokenBtn, tokenEntry)

	resultsContent := container.NewVBox(
		urlRow,
		tokenRow,
		container.NewCenter(copyBothBtn),
	)

	card := widget.NewCard("Upload Complete", "", resultsContent)

	cardContainer := container.NewVBox(card)
	cardContainer.Hide()

	state.uploadComplete.AddListener(binding.NewDataListener(func() {
		complete, _ := state.uploadComplete.Get()
		if complete {
			cardContainer.Show()
		} else {
			cardContainer.Hide()
		}
	}))

	return cardContainer
}

func createStatusBar(state *AppState) fyne.CanvasObject {
	statusLabel := widget.NewLabelWithData(state.statusMessage)
	statusLabel.Importance = widget.LowImportance

	separator := widget.NewSeparator()

	return container.NewVBox(
		separator,
		container.NewPadded(statusLabel),
	)
}

func handleFileSelection(state *AppState, path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		showNotification(state, "Could not read file information", true)
		return
	}

	state.filePath.Set(path)
	state.fileName.Set(filepath.Base(path))
	state.fileSize.Set(formatFileSize(fileInfo.Size()))
	state.uploadComplete.Set(false)
	state.urlBind.Set("")
	state.tokenBind.Set("")
	state.statusMessage.Set(fmt.Sprintf("Ready to upload: %s", filepath.Base(path)))
}

func chooseFile(state *AppState) {
	fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			showNotification(state, "Error opening file dialogue", true)
			return
		}

		if file == nil {
			return
		}

		handleFileSelection(state, file.URI().Path())
	}, state.window)

	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func performUpload(state *AppState) {
	path, _ := state.filePath.Get()
	if path == "" {
		showNotification(state, "No file selected", true)
		return
	}

	state.isUploading.Set(true)
	state.uploadProgress.Set(0)
	state.statusMessage.Set("Preparing upload...")

	go func() {
		// Simulate progress updates
		go func() {
			for i := 0; i < 90; i += 10 {
				time.Sleep(100 * time.Millisecond)
				uploading, _ := state.isUploading.Get()
				if !uploading {
					return
				}
				state.uploadProgress.Set(float64(i) / 100.0)
			}
		}()

		state.statusMessage.Set("Uploading file...")

		fileStruc, err := os.Open(path)
		if err != nil {
			state.isUploading.Set(false)
			showNotification(state, fmt.Sprintf("Error opening file: %v", err), true)
			return
		}
		defer fileStruc.Close()

		resp, err := state.api.UploadFile(context.TODO(), waifuMod.WaifuvaultPutOpts{
			File: fileStruc,
		})

		if err != nil {
			state.isUploading.Set(false)
			state.uploadProgress.Set(0)
			showNotification(state, fmt.Sprintf("Upload failed: %v", err), true)
			return
		}

		state.uploadProgress.Set(1.0)
		state.urlBind.Set(resp.URL)
		state.tokenBind.Set(resp.Token)
		state.isUploading.Set(false)
		state.uploadComplete.Set(true)
		state.statusMessage.Set("Upload successful!")
	}()
}

func showNotification(state *AppState, message string, isError bool) {
	if isError {
		dialog.ShowError(errors.New(message), state.window)
		state.statusMessage.Set(fmt.Sprintf("Error: %s", message))
	} else {
		dialog.ShowInformation("Success", message, state.window)
		state.statusMessage.Set(message)
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
