package main

import (
	"fmt"
	"net/http"
	discordrpc "qusic/discord-rpc"
	"qusic/logger"
	"qusic/lyrics"
	pl "qusic/player"
	"qusic/widgets"
	"qusic/youtube"
	"time"

	"fyne.io/fyne/v2"
	a "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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
var app = a.NewWithID("il.oq.qusic")

var tabs *container.AppTabs
var settingsButton *widget.Button
var rpc = discordrpc.Client{ClientID: "1233164951342813275"}
var fulld *widget.RichText
var apprunning bool
var bottom *fyne.Container

var (
	back, next *widget.Button
	pause      *widgets.RoundedButton
)

func init() {
	player.Initialize()
	app.Settings().SetTheme(&myTheme{})
}

func homePage() fyne.CanvasObject {
	return container.NewWithoutLayout()
}

var syncedLyrics []lyrics.SyncedLyric

var searchContent = (fyne.CanvasObject)(container.NewWithoutLayout())

var songProgressSlider *widget.Slider

func durString(dur time.Duration) string {
	return fmt.Sprintf("%0d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
}

func setPlayedSong(song *pl.Song, w fyne.Window) {
	logger.Infof("Now played song: %s (%s)", song.Name, song.URL)
	err := song.FetchSongInfo()
	if err != nil {
		logger.Errorf("No lyrics for %s:%v", song.Name, err)
	} else {
		logger.Infof("Fetched lyrics for %s", song.Name)
	}

	syncedLyrics = song.Lyrics.SyncedLyrics

	lyricsTxt.Segments = make([]widget.RichTextSegment, len(syncedLyrics))
	for i, lyric := range syncedLyrics {
		var seg = new(widget.TextSegment)
		seg.Style.SizeName = theme.SizeNameHeadingText
		seg.Text = lyric.Lyric
		seg.Style.TextStyle.Bold = false
		lyricsTxt.Segments[i] = seg
	}
	lyricsTxt.Refresh()
	lyricsScroll.ScrollToTop()

	image := song.Thumbnails[0]
	d, _ := http.Get(image.URL)
	img := canvas.NewImageFromReader(d.Body, song.Name)

	songProgressSlider.Max = float64(song.Duration / time.Millisecond)
	songProgressSlider.Enable()

	fulld.Segments[0].(*widget.TextSegment).Text = durString(song.Duration)
	fulld.Segments[0].(*widget.TextSegment).Style.ColorName = theme.ColorNameForeground
	fulld.Refresh()
	back.Enable()
	pause.Enable()
	next.Enable()

	songinfo := &widgets.SongInfo{
		Name:   song.Name,
		Artist: song.Artists[0].Name,
		Image:  img,
	}
	bottom.Objects[0] = songinfo
	bottom.Refresh()
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
		logger.Inff("Searching youtube music (query:\"%s\"): ", s)
		songs, e := client.SearchSongs(s)
		videos, e1 := client.SearchVideos(s)
		fmt.Printf("%d songs (%v), %d videos (%v)\n", len(songs), e, len(videos), e1)

		results := [2][]youtube.MusicSearchResult{songs, videos}
		var forms = [2]fyne.CanvasObject{
			layout.NewSpacer(), layout.NewSpacer(),
		}

		for i, res := range results {
			if len(res) == 0 {
				continue
			}
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
					Artist:         song.Authors[0].Name,
					Image:          img,
					DurationString: durString(song.Duration),
					OnTapped: func() {
						go func() {
							so := player.YoutubeMusicSong(song)
							logger.Inff("Playing song %s: ", so.Name)
							err := player.PlayNow(so)
							fmt.Println(err)
							if err != nil {
								return
							}
							setPlayedSong(so, w)
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

func main() {
	logger.Info("Qusic [ made by oq ]")

	app.SetIcon(resourceQusicPng)
	var window fyne.Window

	app.Lifecycle().SetOnStarted(func() {
		logger.Inf("Connecting to Discord RPC: ")
		fmt.Println(rpc.Connect())

		logger.Info("Launching app")
		apprunning = true
		window = app.NewWindow("Qusic")
		window.SetCloseIntercept(func() {
			window.Hide()
		})

		tabs = container.NewAppTabs(
			container.NewTabItemWithIcon("Home", theme.HomeIcon(), homePage()),
			container.NewTabItemWithIcon("Search", theme.SearchIcon(), searchPage(window)),
			container.NewTabItem("Lyrics", lyricsPage()),
		)

		tabs.SetTabLocation(container.TabLocationLeading)

		back = &widget.Button{Icon: theme.MediaSkipPreviousIcon(), Importance: widget.LowImportance}
		back.Disable()

		pause = &widgets.RoundedButton{
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
		pause.Disable()
		next = &widget.Button{Icon: theme.MediaSkipNextIcon(), Importance: widget.LowImportance}
		next.Disable()

		settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			settings(window)
		})
		settingsButton.Importance = widget.LowImportance

		songProgressSlider = widget.NewSlider(0, 0)
		songProgressSlider.Disable()

		var p *widget.RichText
		p, fulld = widget.NewRichText(&widget.TextSegment{Text: "0:00", Style: widget.RichTextStyle{ColorName: theme.ColorNameDisabled}}), widget.NewRichText(&widget.TextSegment{Text: "-:--", Style: widget.RichTextStyle{ColorName: theme.ColorNameDisabled}})
		prevf := 0.0

		songProgressSlider.OnChanged = func(f float64) {
			passed := time.Duration(f) * time.Millisecond

			p.Segments[0].(*widget.TextSegment).Text = durString(passed)
			p.Segments[0].(*widget.TextSegment).Style.ColorName = theme.ColorNameForeground
			p.Refresh()
			if f-prevf > 1000 {
				player.Seek(int(f / 1000))
			}
			prevf = f
		}

		progress := container.NewBorder(nil, nil, p, fulld, songProgressSlider)
		control := container.NewGridWithRows(2, container.NewHBox(layout.NewSpacer(), back, pause, next, layout.NewSpacer()), progress)

		bottom = container.NewGridWithColumns(3, layout.NewSpacer(), control)

		window.SetContent(container.NewBorder(nil, bottom, nil, container.NewVBox(settingsButton), tabs))
		resolution := screenresolution.GetPrimary()
		window.Resize(fyne.NewSize(float32(resolution.Width)/1.5, float32(resolution.Height)/1.5))
		window.Show()

		tick := time.NewTicker(time.Millisecond)
		for {
			if !player.Playing() {
				continue
			}
			passed, _ := player.TimePosition(false)
			select {
			case <-tick.C:
				if songProgressSlider == nil {
					continue
				}
				cs := player.CurrentSong()
				rpc.SetActivity(discordrpc.Activity{
					Type:    discordrpc.Listening,
					Details: cs.Name,
					State:   "By " + cs.Artists[0].Name,
					Timestamps: discordrpc.ActivityTimestamps{
						Start: int(time.Now().UnixMilli()) - int(passed/time.Millisecond),
						End:   int(time.Now().UnixMilli()) + int(cs.Duration/time.Millisecond) - int(passed/time.Millisecond),
					},
				})
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

					if lyric.Index != 0 {
						lyricsScroll.Scrolled(&fyne.ScrollEvent{
							Scrolled: fyne.NewDelta(0, -30),
						})
					}
				}
			}
		}
	})

	if desk, ok := app.(desktop.App); ok {
		desk.SetSystemTrayMenu(fyne.NewMenu("Qusic", fyne.NewMenuItem("Show", func() {
			window.Show()
		})))
	}

	app.Run()
}
