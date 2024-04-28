package main

import (
	"fmt"
	"net/http"
	"qusic/lyrics"
	pl "qusic/player"
	"qusic/widgets"
	"qusic/youtube"
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

//var _ = godotenv.Load()
//var id = os.Getenv("SPOTIFY_ID")
//var secret = os.Getenv("SPOTIFY_SECRET")

var client = new(youtube.MusicClient)
var player = pl.New()

func init() {
	player.Initialize()
}

func homePage() fyne.CanvasObject {
	return container.NewWithoutLayout()
}

var syncedLyrics []lyrics.SyncedLyric

var searchContent = (fyne.CanvasObject)(container.NewWithoutLayout())

var songProgressSlider *widget.Slider

func durString(dur time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
}

func setPlayedSong(song *pl.Song, w fyne.Window) {
	song.FetchSongInfo()

	syncedLyrics = song.SongInfo.SyncedLyrics

	lyricsTxt.Segments = make([]widget.RichTextSegment, len(syncedLyrics))
	for i, lyric := range syncedLyrics {
		var seg = new(widget.TextSegment)
		seg.Style.SizeName = theme.SizeNameHeadingText
		seg.Text = lyric.Lyric
		seg.Style.TextStyle.Bold = false
		lyricsTxt.Segments[i] = seg
	}
	lyricsTxt.Refresh()

	image := song.Thumbnails[0]
	d, _ := http.Get(image.URL)
	img := canvas.NewImageFromReader(d.Body, song.Title)

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

	songProgressSlider = widget.NewSlider(0, float64(song.Duration/time.Millisecond))

	p, f := widget.NewLabel("00:00"), widget.NewLabel(durString(song.Duration))
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
		Name:   song.Title,
		Artist: song.Author.Name,
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
		songs, _ := client.SearchSongs(s)
		videos, _ := client.SearchVideos(s)

		results := [2][]youtube.MusicSearchResult{songs, videos}
		var forms [2]*fyne.Container

		for i, res := range results {
			txt := "Songs"
			if i == 1 {
				txt = "Videos"
			}
			form := container.NewVBox(widget.NewRichTextFromMarkdown("# " + txt))
			songsc := container.NewVBox()

			for _, s := range res {
				song := s
				image := song.Thumbnails[0]
				d, _ := http.Get(image.URL)
				img := canvas.NewImageFromReader(d.Body, song.Title)
				img.SetMinSize(fyne.NewSize(48, 48))
				songsc.Add(&widgets.SongResult{
					Name:           song.Title,
					Artist:         song.Author.Name,
					Image:          img,
					DurationString: durString(song.Duration),
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
			forms[i] = form
		}

		grid := container.NewGridWithColumns(2, container.NewVScroll(forms[0]), container.NewVScroll(forms[1]))
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
				if len(syncedLyrics) == 0 {
					continue
				}
				lyric := syncedLyrics[0]
				if lyric.At <= passed {
					lyricsTxt.Segments[lyric.Index].(*widget.TextSegment).Style.TextStyle.Bold = true
					lyricsTxt.Refresh()
					syncedLyrics = syncedLyrics[1:]
				}
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
