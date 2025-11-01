package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

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

type FileItem struct {
	path  string
	name  string
	size  int64
	url   string
	token string
}

type AppState struct {
	api              waifuMod.Waifuvalt
	window           fyne.Window
	files            []FileItem
	filesContainer   *fyne.Container
	resultsContainer *fyne.Container
	uploadButton     *widget.Button
	isUploading      binding.Bool
	uploadProgress   binding.Float
	uploadComplete   binding.Bool
	statusMessage    binding.String
}

func main() {
	api := waifuVault.NewWaifuvaltApi(http.Client{})

	waifuVaultApp := app.NewWithID("moe.waifuvault.desktop")
	w := waifuVaultApp.NewWindow("WaifuVault - File Sharing Made Easy")
	w.Resize(fyne.NewSize(650, 450))
	w.CenterOnScreen()
	w.SetFixedSize(false)

	state := &AppState{
		api:              api,
		window:           w,
		files:            []FileItem{},
		filesContainer:   container.NewVBox(),
		resultsContainer: container.NewVBox(),
		isUploading:      binding.NewBool(),
		uploadProgress:   binding.NewFloat(),
		uploadComplete:   binding.NewBool(),
		statusMessage:    binding.NewString(),
	}

	state.statusMessage.Set("Ready to upload")

	content := buildUI(state)
	w.SetContent(content)

	// Handle drag and drop
	w.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}
		for _, uri := range uris {
			addFileToList(state, uri.Path())
		}
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
	header := container.NewBorder(
		nil, nil,
		widget.NewLabel("Selected Files"),
		widget.NewButton("Clear All", func() {
			state.files = []FileItem{}
			state.filesContainer.Objects = []fyne.CanvasObject{}
			state.filesContainer.Refresh()
			state.uploadComplete.Set(false)
			state.statusMessage.Set("Ready to upload")

			// Update button state
			updateUploadButtonState(state)
		}),
	)

	scrollContainer := container.NewVScroll(state.filesContainer)
	scrollContainer.SetMinSize(fyne.NewSize(0, 150))

	card := widget.NewCard("", "", container.NewBorder(header, nil, nil, nil, scrollContainer))

	return card
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
	uploadBtn := widget.NewButtonWithIcon("Upload Files", theme.UploadIcon(), func() {
		performUpload(state)
	})
	uploadBtn.Importance = widget.HighImportance
	uploadBtn.Disable()

	state.uploadButton = uploadBtn

	state.isUploading.AddListener(binding.NewDataListener(func() {
		updateUploadButtonState(state)
	}))

	return container.NewCenter(uploadBtn)
}

func updateUploadButtonState(state *AppState) {
	if state.uploadButton == nil {
		return
	}

	uploading, _ := state.isUploading.Get()
	if uploading {
		state.uploadButton.Disable()
	} else {
		if len(state.files) > 0 {
			state.uploadButton.Enable()
		} else {
			state.uploadButton.Disable()
		}
	}
}

func createResultsCard(state *AppState) fyne.CanvasObject {
	scrollContainer := container.NewVScroll(state.resultsContainer)
	scrollContainer.SetMinSize(fyne.NewSize(0, 200))

	card := widget.NewCard("Upload Results", "", scrollContainer)

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

func addFileToList(state *AppState, path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		state.statusMessage.Set("Could not read file information")
		return
	}

	// Check if file already exists
	for _, f := range state.files {
		if f.path == path {
			state.statusMessage.Set("File already in list")
			return
		}
	}

	fileItem := FileItem{
		path: path,
		name: filepath.Base(path),
		size: fileInfo.Size(),
	}

	state.files = append(state.files, fileItem)
	updateFilesList(state)
	state.uploadComplete.Set(false)
	state.statusMessage.Set(fmt.Sprintf("%d file(s) ready to upload", len(state.files)))

	// Update button state
	updateUploadButtonState(state)
}

func updateFilesList(state *AppState) {
	state.filesContainer.Objects = []fyne.CanvasObject{}

	for i, file := range state.files {
		idx := i
		fileIcon := widget.NewIcon(theme.FileIcon())

		nameLabel := widget.NewLabel(file.name)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		sizeLabel := widget.NewLabel(formatFileSize(file.size))
		sizeLabel.TextStyle = fyne.TextStyle{Italic: true}

		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			removeFileFromList(state, idx)
		})
		removeBtn.Importance = widget.LowImportance

		fileRow := container.NewBorder(
			nil, nil,
			container.NewHBox(fileIcon, container.NewVBox(nameLabel, sizeLabel)),
			removeBtn,
		)

		state.filesContainer.Add(fileRow)
	}

	state.filesContainer.Refresh()
}

func removeFileFromList(state *AppState, index int) {
	if index >= 0 && index < len(state.files) {
		state.files = append(state.files[:index], state.files[index+1:]...)
		updateFilesList(state)

		if len(state.files) == 0 {
			state.statusMessage.Set("Ready to upload")
		} else {
			state.statusMessage.Set(fmt.Sprintf("%d file(s) ready to upload", len(state.files)))
		}

		// Update button state
		updateUploadButtonState(state)
	}
}

func chooseFile(state *AppState) {
	dialog.ShowFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			state.statusMessage.Set("Error opening file")
			return
		}

		if file == nil {
			return
		}

		addFileToList(state, file.URI().Path())
	}, state.window)
}

func performUpload(state *AppState) {
	if len(state.files) == 0 {
		state.statusMessage.Set("No files selected")
		return
	}

	state.isUploading.Set(true)
	state.uploadProgress.Set(0)

	go func() {
		totalFiles := len(state.files)
		state.statusMessage.Set(fmt.Sprintf("Uploading %d file(s) in parallel...", totalFiles))
		var wg sync.WaitGroup
		var completedCount int32
		var failedCount int32
		var mu sync.Mutex

		// Launch a goroutine for each file
		for i := range state.files {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				fileStruc, err := os.Open(state.files[index].path)
				if err != nil {
					atomic.AddInt32(&failedCount, 1)
					state.statusMessage.Set(fmt.Sprintf("Error opening file %s: %v", state.files[index].name, err))
					return
				}
				defer fileStruc.Close()

				resp, err := state.api.UploadFile(context.TODO(), waifuMod.WaifuvaultPutOpts{
					File: fileStruc,
				})

				if err != nil {
					atomic.AddInt32(&failedCount, 1)
					state.statusMessage.Set(fmt.Sprintf("Upload failed for %s: %v", state.files[index].name, err))
					return
				}

				// Store URL and token in the file item (with mutex to avoid race conditions)
				mu.Lock()
				state.files[index].url = resp.URL
				state.files[index].token = resp.Token
				mu.Unlock()

				// Update progress
				completed := atomic.AddInt32(&completedCount, 1)
				progress := float64(completed) / float64(totalFiles)
				state.uploadProgress.Set(progress)
				state.statusMessage.Set(fmt.Sprintf("Uploaded %d of %d files", completed, totalFiles))
			}(i)
		}

		// Wait for all uploads to complete
		wg.Wait()

		state.uploadProgress.Set(1.0)
		updateResultsList(state)
		state.isUploading.Set(false)
		state.uploadComplete.Set(true)

		failed := atomic.LoadInt32(&failedCount)
		if failed > 0 {
			state.statusMessage.Set(fmt.Sprintf("Upload complete: %d succeeded, %d failed", totalFiles-int(failed), failed))
		} else {
			state.statusMessage.Set(fmt.Sprintf("Successfully uploaded all %d file(s)!", totalFiles))
		}
	}()
}

func updateResultsList(state *AppState) {
	state.resultsContainer.Objects = []fyne.CanvasObject{}

	for _, file := range state.files {
		if file.url == "" {
			continue
		}

		fileIcon := widget.NewIcon(theme.ConfirmIcon())

		nameLabel := widget.NewLabel(file.name)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		urlLabel := widget.NewLabel(file.url)
		urlLabel.Wrapping = fyne.TextWrapBreak

		tokenLabel := widget.NewLabel(file.token)
		tokenLabel.Wrapping = fyne.TextWrapBreak

		copyUrlBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(url string) func() {
			return func() {
				state.window.Clipboard().SetContent(url)
				state.statusMessage.Set("URL copied!")
			}
		}(file.url))
		copyUrlBtn.Importance = widget.LowImportance

		copyTokenBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(token string) func() {
			return func() {
				state.window.Clipboard().SetContent(token)
				state.statusMessage.Set("Token copied!")
			}
		}(file.token))
		copyTokenBtn.Importance = widget.LowImportance

		copyBothBtn := widget.NewButtonWithIcon("Copy Both", theme.ContentCopyIcon(), func(url, token string) func() {
			return func() {
				both := fmt.Sprintf("URL: %s\nToken: %s", url, token)
				state.window.Clipboard().SetContent(both)
				state.statusMessage.Set("URL and token copied!")
			}
		}(file.url, file.token))
		copyBothBtn.Importance = widget.LowImportance

		urlHeader := widget.NewLabel("URL:")
		urlHeader.TextStyle = fyne.TextStyle{Bold: true}

		tokenHeader := widget.NewLabel("Token:")
		tokenHeader.TextStyle = fyne.TextStyle{Bold: true}

		urlRow := container.NewBorder(nil, nil, nil, copyUrlBtn, urlLabel)
		tokenRow := container.NewBorder(nil, nil, nil, copyTokenBtn, tokenLabel)

		fileResult := container.NewVBox(
			container.NewHBox(fileIcon, nameLabel),
			widget.NewSeparator(),
			urlHeader,
			urlRow,
			tokenHeader,
			tokenRow,
			container.NewCenter(copyBothBtn),
		)

		card := widget.NewCard("", "", fileResult)
		state.resultsContainer.Add(card)
	}

	state.resultsContainer.Refresh()
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
