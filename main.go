package main

import (
	"fmt"
	"net/http"
	"os"
	pl "qusic/player"
	"qusic/spotify"
	"qusic/widgets"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fstanis/screenresolution"
	"github.com/joho/godotenv"
)

var _ = godotenv.Load()
var id = os.Getenv("SPOTIFY_ID")
var secret = os.Getenv("SPOTIFY_SECRET")

var client = spotify.New(id, secret)
var player = pl.New(client)

func init() {
	player.Initialize()
}

func homePage() fyne.CanvasObject {
	return container.NewWithoutLayout()
}

var searchContent = (fyne.CanvasObject)(container.NewWithoutLayout())

var songProgressSlider *widget.Slider

func durString(dur time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
}

func setPlayedSong(song *pl.Song, w fyne.Window) {
	song.FetchSongInfo()

	image := song.Album.Images[2]
	d, _ := http.Get(image.URL)
	img := canvas.NewImageFromReader(d.Body, song.Name)

	back := widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), nil)
	back.Importance = widget.LowImportance
	pause := &widgets.RoundedButton{
		Icon: theme.MediaPauseIcon(),
	}
	pause.OnTapped = func() {
		player.PauseCycle()
		if player.Paused() {
			pause.SetIcon(theme.MediaPlayIcon())
		} else {
			pause.SetIcon(theme.MediaPauseIcon())
		}
	}
	next := widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nil)
	next.Importance = widget.LowImportance

	songProgressSlider = widget.NewSlider(0, float64(song.DurationMS))

	full := time.Millisecond * time.Duration(song.DurationMS)

	p, f := widget.NewLabel("00:00"), widget.NewLabel(durString(full))
	prevf := 0.0

	songProgressSlider.OnChanged = func(f float64) {
		passed := time.Duration(f) * time.Millisecond

		p.SetText(durString(passed))
		if f-prevf > 1000 {
			player.Seek(int(f / 1000))
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
	searchButton := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		if searchBar.Text == "" {
			return
		}
		searchBar.OnSubmitted(searchBar.Text)
	})

	border := container.NewBorder(container.NewGridWithColumns(3, layout.NewSpacer(), container.NewBorder(nil, nil, nil, searchButton, searchBar)), nil, nil, nil, searchContent)

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
					go func() {
						so := player.Song(song)
						setPlayedSong(so, w)
						player.PlayNow(so)
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
	a.Settings().SetTheme(&myTheme{})
	w := a.NewWindow("Qusic")
	tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Home", theme.HomeIcon(), homePage()),
		container.NewTabItemWithIcon("Search", theme.SearchIcon(), searchPage(w)),
		container.NewTabItem("Lyrics", lyricsPage()),
	)
	tick := time.NewTicker(time.Millisecond)
	go func() {
		for {
			if !player.Playing() {
				continue
			}
			passed, _ := player.TimePosition(false)
			select {
			case <-tick.C:
				songProgressSlider.SetValue(float64(passed / time.Millisecond))
			default:
				syncedLyrics := player.CurrentSong().SongInfo.SyncedLyrics

				if len(lyricsTxt.Segments) != len(syncedLyrics) {
					lyricsTxt.Segments = make([]widget.RichTextSegment, len(syncedLyrics))
				}

				for i, lyric := range syncedLyrics {
					var seg = new(widget.TextSegment)
					seg.Style.SizeName = theme.SizeNameHeadingText
					if lyric.At <= passed {
						seg.Style.TextStyle.Bold = true
					}
					seg.Text = lyric.Lyric
					lyricsTxt.Segments[i] = seg
				}
				lyricsTxt.Refresh()
			}
		}
	}()

	tabs.SetTabLocation(container.TabLocationLeading)

	border0 = container.NewBorder(nil, container.NewCenter(widget.NewRichTextFromMarkdown("# Nothing is playing...")), nil, nil, tabs)
	w.SetContent(border0)
	resolution := screenresolution.GetPrimary()
	w.Resize(fyne.NewSize(float32(resolution.Width), float32(resolution.Height)))
	w.ShowAndRun()
}
