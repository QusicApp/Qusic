package main

import (
	"fmt"
	"image/color"
	"net/http"
	"qusic/spotify"
	"qusic/widgets"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fstanis/screenresolution"
)

func homePage() fyne.CanvasObject {
	return container.NewWithoutLayout()
}

var searchContent = (fyne.CanvasObject)(container.NewWithoutLayout())

// var songProgress *widget.ProgressBar
var songProgressSlider *widget.Slider

func durString(dur time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
}

func setPlayedSong(song spotify.TrackObject, w fyne.Window) {
	<-songChanged
	song.FetchSongInfo()
	currentSong = song

	image := song.Album.Images[2]
	d, _ := http.Get(image.URL)
	img := canvas.NewImageFromReader(d.Body, song.Name)

	back := widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), nil)
	back.Importance = widget.LowImportance
	//pause := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil)
	pause := &widgets.RoundedButton{
		Icon:  theme.MediaPauseIcon(),
		Color: color.White,
	}
	pause.OnTapped = func() {
		pauseResume()
		if paused {
			pause.Icon = theme.MediaPlayIcon()
		} else {
			pause.Icon = theme.MediaPauseIcon()
		}
		pause.Refresh()
	}
	next := widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nil)
	next.Importance = widget.LowImportance

	songProgressSlider = widget.NewSlider(0, float64(currentSong.DurationMS))

	full := time.Millisecond * time.Duration(currentSong.DurationMS)

	p, f := widget.NewLabel("00:00"), widget.NewLabel(durString(full))
	prevf := 0.0

	songProgressSlider.OnChanged = func(f float64) {
		passed := time.Duration(f) * time.Millisecond

		p.SetText(durString(passed))
		if f-prevf > 1000 {
			seek(int(f / 1000))
		}
		prevf = f
	}

	songinfo := &widgets.SongInfo{
		Name:   song.Name,
		Artist: song.Artists[0].Name,
		Image:  img,
	}

	progress := container.NewBorder(nil, nil, p, f, songProgressSlider)
	control := container.NewGridWithRows(2, container.NewHBox(layout.NewSpacer(), back, pause, next, layout.NewSpacer()), progress)

	x := container.NewGridWithColumns(3, container.NewHBox(songinfo), control)

	border0 = container.NewBorder(nil, x, nil, nil, tabs)
	w.SetContent(border0)
}

func searchPage(w fyne.Window) fyne.CanvasObject {
	searchBar := widget.NewEntry()
	searchBar.SetPlaceHolder("What do you want to play?")

	border := container.NewBorder(container.NewGridWithColumns(3, layout.NewSpacer(), searchBar), nil, nil, nil, searchContent)

	searchBar.OnSubmitted = func(s string) {
		res, _ := client.Search(s, spotify.QueryAll, "", nil, nil, false)
		form := container.NewVBox(widget.NewRichTextFromMarkdown("# Songs"))
		songsc := container.NewVBox()

		for _, s := range res.Tracks.Items {
			song := s
			image := song.Album.Images[1]
			d, _ := http.Get(image.URL)
			img := canvas.NewImageFromReader(d.Body, song.Name)
			img.SetMinSize(fyne.NewSize(48, 48))
			songsc.Add(&widgets.SongResult{
				Name:           song.Name,
				Artist:         song.Artists[0].Name,
				Image:          img,
				DurationString: durString(time.Duration(song.DurationMS) * time.Millisecond),
				OnTapped: func() {
					playSong(song)
					setPlayedSong(song, w)

					tick := time.NewTicker(time.Millisecond)
					go func() {
						for {
							select {
							case <-songChanged:
								return
							case <-tick.C:
								// in seconds
								passed, _ := strconv.ParseFloat(player.GetPropertyString("time-pos"), 64)
								songProgressSlider.SetValue(passed * 1000)
							}
						}
					}()
				},
			})
		}
		form.Add(songsc)

		grid := container.NewGridWithColumns(2, layout.NewSpacer(), container.NewScroll(form))
		searchContent = grid
		border.Objects[0] = searchContent
		border.Refresh()
	}
	return border
}

var border0 *fyne.Container
var tabs *container.AppTabs

func main() {
	a := app.NewWithID("il.oq.qusic")
	w := a.NewWindow("Qusic")
	tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Home", theme.HomeIcon(), homePage()),
		container.NewTabItemWithIcon("Search", theme.SearchIcon(), searchPage(w)),
		container.NewTabItem("Lyrics", lyricsPage()),
	)
	tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Lyrics" {
			ti.Content = lyricsPage()
			tabs.Refresh()
		}
	}
	tabs.SetTabLocation(container.TabLocationLeading)

	border0 = container.NewBorder(nil, container.NewCenter(widget.NewRichTextFromMarkdown("# Nothing is playing...")), nil, nil, tabs)
	w.SetContent(border0)
	resolution := screenresolution.GetPrimary()
	w.Resize(fyne.NewSize(float32(resolution.Width), float32(resolution.Height)))
	w.ShowAndRun()
}
