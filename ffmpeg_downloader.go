package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// wip

func ffmpegDownloader() {
	w := app.Driver().(desktop.Driver).CreateSplashWindow()
	w.SetPadded(true)

	pathEntry := widget.NewEntry()
	pathEntry.SetText(defaultPath())

	progress := widget.NewProgressBar()
	progress.Hide()

	installButton := &widget.Button{
		Importance: widget.HighImportance,
		Text:       "Install",
	}
	installButton.OnTapped = func() {
		pathEntry.Disable()
		installButton.Disable()
		installButton.SetText("Installing...")
		fmt.Println(downloadTo(pathEntry.Text))
	}
	border := container.NewBorder(
		widget.NewRichText(
			&widget.TextSegment{Style: widget.RichTextStyleHeading, Text: "Qusic FFMPEG Installer"},
			&widget.TextSegment{Text: "Qusic needs to install ffmpeg for the app to work."},
		),
		container.NewGridWithColumns(3, layout.NewSpacer(),
			container.NewPadded(installButton),
		), nil, nil,
		container.NewPadded(
			container.NewVBox(
				widget.NewLabel("Installation path:"),
				progress,
				pathEntry,
			)),
	)
	w.SetContent(
		container.NewStack(
			container.NewHBox(layout.NewSpacer(), container.NewVBox(&widget.Button{
				Importance: widget.DangerImportance,
				Icon:       theme.WindowCloseIcon(),
				OnTapped: func() {
					app.Quit()
				},
			}, layout.NewSpacer())),
			border,
		),
	)

	size := w.Content().MinSize()
	size.Width *= 1.2
	size.Height *= 1.2
	w.Resize(size)

	w.ShowAndRun()
}

func defaultPath() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\Program Files\\Qusic"
	default:
		return "/etc/qusic"
	}
}

func fileName() string {
	switch runtime.GOOS {
	case "windows":
		return "\\ffmpeg.zip"
	case "darwin":
		return "/ffmpeg.zip"
	default:
		return "/ffmpeg.tar.xz"
	}
}

func downloadTo(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("No access. Please run the app in administrator/elevated mode.")
		}
		return err
	}
	url := downloadLink()
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	file, err := os.Create(path + fileName())
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("No access. Please run the app in administrator/elevated mode.")
		}
		return err
	}
	i, err := file.ReadFrom(res.Body)
	if err != nil {
		return err
	}
	file.Close()

	if res.ContentLength != i {
		return fmt.Errorf("Corrupted file.")
	}
	return nil
}

func downloadLink() string {
	switch runtime.GOOS {
	case "windows":
		return "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
	case "darwin":
		return "https://evermeet.cx/ffmpeg/getrelease/zip"
	case "linux":
		return "wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz"
	default:
		return ""
	}
}
