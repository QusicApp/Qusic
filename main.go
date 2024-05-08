package main

import (
	"fmt"
	"image/color"
	"math"
	"net/http"
	discordrpc "qusic/discord-rpc"
	"qusic/logger"
	"qusic/lyrics"
	pl "qusic/player"
	"strings"

	"qusic/widgets"
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

var youtubeSource pl.YouTubeMusicSource
var spotifySource = pl.NewSpotifySource()

var player = pl.New(youtubeSource)
var app = a.NewWithID("il.oq.qusic")

var tabs *container.AppTabs
var settingsButton *widget.Button
var rpc = discordrpc.Client{ClientID: "1233164951342813275"}
var fulld *widget.RichText
var apprunning bool
var bottom *fyne.Container

var (
	back, next *widgets.Button
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
	if dur > time.Hour {
		return fmt.Sprintf("%0d:%02d:%02d", int(dur.Hours())%60, int(dur.Minutes())%60, int(dur.Seconds())%60)
	} else {
		return fmt.Sprintf("%0d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
	}
}

func setPlayedSong(song *pl.Song, w fyne.Window) {
	logger.Infof("Now played song: %s (%s)", song.Name, song.URL)

	source := preferences.String("lyrics.source")
	var err error
	switch source {
	case "lrclib":
		song.Lyrics, err = lyrics.GetSongLRCLIB(song.Name, song.Artists[0].Name, song.Album.Name, song.Duration, false)
	case "ytmusic":
		song.Lyrics, err = lyrics.GetSongYTMusic(song.Video.ID)
	case "genius":
		song.Lyrics, err = lyrics.GetSongGenius(song.Artists[0].Name, song.Name)
	}
	if err != nil {
		logger.Errorf("No lyrics for %s<source:%s,album:%s,artist:%s,duration:%s>:%v", song.Name, source, song.Album.Name, song.Artists[0].Name, song.Duration, err)
	} else {
		logger.Infof("Fetched lyrics for %s (source: %s)", song.Name, source)
	}

	syncedLyrics = song.Lyrics.SyncedLyrics

	if len(syncedLyrics) == 0 {
		lines := strings.Split(song.Lyrics.PlainLyrics, "\n")
		lyricsTxt.Segments = make([]widget.RichTextSegment, len(lines))
		for i, line := range lines {
			var seg = new(widget.TextSegment)
			seg.Style.SizeName = theme.SizeNameHeadingText
			seg.Text = line
			seg.Style.TextStyle.Bold = false

			lyricsTxt.Segments[i] = seg
		}
	} else {
		lyricsTxt.Segments = make([]widget.RichTextSegment, len(syncedLyrics))
		for i, lyric := range syncedLyrics {
			//segment := i
			var seg = new(widget.TextSegment)
			seg.Style.SizeName = theme.SizeNameHeadingText
			seg.Text = lyric.Lyric
			seg.Style.TextStyle.Bold = false
			/*seg.OnTapped = func() {
				player.Seek(int(song.Lyrics.SyncedLyrics[segment].At / time.Second))

				if segment < syncedLyrics[0].Index {
					for _, s := range lyricsTxt.Segments[:syncedLyrics[0].Index] {
						s.(*widget.TextSegment).Style.TextStyle.Bold = false
					}
				}
				for _, s := range lyricsTxt.Segments[:segment] {
					s.(*widget.TextSegment).Style.TextStyle.Bold = true
				}
				syncedLyrics = syncedLyrics[segment:]
			}*/
			lyricsTxt.Segments[i] = seg
		}
	}
	lyricsTxt.Segments = append(lyricsTxt.Segments, &widget.TextSegment{
		Style: widget.RichTextStyle{SizeName: theme.SizeNameSubHeadingText},
		Text:  "Source: " + song.Lyrics.LyricSource,
	})
	lyricsTxt.Refresh()

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

func main() {
	logger.Info("Qusic [ made by oq ]")
	if preferences.String("source") == "spotify" {
		player.Source = spotifySource
	}

	app.SetIcon(resourceQusicPng)
	var window fyne.Window

	app.Lifecycle().SetOnStarted(func() {
		if preferences.Bool("discord_rpc") {
			logger.Inf("Connecting to Discord RPC: ")
			logger.Println(rpc.Connect())
		}

		logger.Info("Launching app")
		apprunning = true
		window = app.NewWindow("Qusic")
		window.SetCloseIntercept(func() {
			if preferences.Bool("hide_app") {
				window.Hide()
			} else {
				app.Quit()
			}
		})

		tabs = container.NewAppTabs(
			container.NewTabItemWithIcon("", theme.HomeIcon(), homePage()),
			container.NewTabItemWithIcon("", theme.SearchIcon(), searchPage(window)),
			container.NewTabItem("Lyrics", lyricsPage()),
		)

		tabs.SetTabLocation(container.TabLocationLeading)

		back = widgets.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
			if p, _ := player.TimePosition(false); p >= time.Second*2 {
				player.Seek(0)
			}
		})
		back.Importance = widget.LowImportance
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
		next = widgets.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nil)
		next.Importance = widget.LowImportance
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
			if math.Abs(f-prevf) > 1000 {
				player.Seek(int(f / 1000))
				cs := player.CurrentSong()
				var in int
				for i, lyric := range cs.Lyrics.SyncedLyrics {
					if lyric.At <= passed {
						in++
					}
					lyricsTxt.Segments[i].(*widget.TextSegment).Style.TextStyle.Bold = lyric.At <= passed
				}
				syncedLyrics = cs.Lyrics.SyncedLyrics[in:]
				lyricsTxt.Refresh()
			}
			prevf = f
		}

		progress := container.NewBorder(nil, nil, p, fulld, songProgressSlider)
		control := container.NewGridWithRows(2, container.NewHBox(layout.NewSpacer(), back, pause, next, layout.NewSpacer()), progress)

		bottom = container.NewGridWithColumns(3, layout.NewSpacer(), control)

		window.SetContent(container.NewBorder(nil,
			container.NewStack(
				canvas.NewRectangle(color.Black),
				container.NewPadded(bottom),
			), nil, container.NewVBox(settingsButton), tabs))
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
				if preferences.Bool("discord_rpc") {
					rpc.SetActivity(discordrpc.Activity{
						Type:    discordrpc.Listening,
						Details: cs.Name,
						State:   "By " + cs.Artists[0].Name,
						Timestamps: discordrpc.ActivityTimestamps{
							Start: int(time.Now().UnixMilli()) - int(passed/time.Millisecond),
							End:   int(time.Now().UnixMilli()) + int(cs.Duration/time.Millisecond) - int(passed/time.Millisecond),
						},
					})
				}
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

					dy := lyricsScroll.Offset.Y - (float32(lyric.Index) * theme.TextHeadingSize())

					lyricsScroll.Scrolled(&fyne.ScrollEvent{
						Scrolled: fyne.NewDelta(0, dy),
					})
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

var resolution = screenresolution.GetPrimary()
