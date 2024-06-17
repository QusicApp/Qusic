package main

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"qusic/constant"
	discordrpc "qusic/discord-rpc"
	"qusic/logger"
	"qusic/lyrics"
	pl "qusic/player"
	"qusic/preferences"
	"qusic/spotify"

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
)

var youtubeSource pl.YouTubeMusicSource
var spotifyClient = spotify.New()
var spotifySource = pl.NewSpotifySource(spotifyClient)

var player = pl.New(youtubeSource)
var app = a.NewWithID("il.oq.qusic")

var tabs *container.AppTabs
var settingsButton *widget.Button
var rpc = discordrpc.Client{ClientID: "1233164951342813275"}
var posd, fulld *widget.RichText
var apprunning bool
var bottom *fyne.Container

var (
	back, next *widgets.Button
	pause      *widgets.RoundedButton
)

func init() {
	app.Settings().SetTheme(&myTheme{})
}

func homePage() fyne.CanvasObject {
	return container.NewWithoutLayout()
}

var syncedLyrics []lyrics.SyncedLyric

var searchContent = (fyne.CanvasObject)(container.NewWithoutLayout())

var songProgressSlider *widget.Slider
var songVolumeSlider *widget.Slider

func durString(dur time.Duration) string {
	if dur > time.Hour {
		return fmt.Sprintf("%0d:%02d:%02d", int(dur.Hours())%60, int(dur.Minutes())%60, int(dur.Seconds())%60)
	} else {
		return fmt.Sprintf("%0d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60)
	}
}

func durStringMS(dur time.Duration) string {
	return fmt.Sprintf("%0d:%02d:%02d", int(dur.Minutes())%60, int(dur.Seconds())%60, int(dur.Milliseconds())%100)
}

var ErrUnsupportedSource = errors.New("unsupported lyrics source")

var (
	searchPaged *widgets.Paged
)

func main() {
	logger.Info("Qusic [ made by oq ]")
	spotifyClient.Cookie_sp_dc = preferences.Preferences.String("spotify.sp_dc")
	if preferences.Preferences.String("source") == "spotify" {
		player.Source = spotifySource
	}

	app.SetIcon(resourceQusicPng)
	var window fyne.Window

	app.Lifecycle().SetOnStarted(func() {
		if preferences.Preferences.Bool("discord_rpc") {
			logger.Inf("Connecting to Discord RPC: ")
			logger.Println(rpc.Connect())
		}

		logger.Info("Launching app")
		apprunning = true
		window = app.NewWindow("Qusic")
		window.SetCloseIntercept(func() {
			if preferences.Preferences.Bool("hide_app") {
				window.Hide()
			} else {
				app.Quit()
			}
		})

		searchPaged = widgets.NewPaged(searchPage(window))
		tabs = container.NewAppTabs(
			container.NewTabItemWithIcon("", theme.HomeIcon(), homePage()),
			container.NewTabItemWithIcon("", theme.SearchIcon(), searchPaged.Container()),
			container.NewTabItem("Lyrics", lyricsPage(window)),
			container.NewTabItem("DJ Mode", djModePage()),
		)

		tabs.DisableIndex(2)

		tabs.SetTabLocation(container.TabLocationLeading)

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

		back = widgets.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
			if player.TimePosition() >= time.Second*5 {
				player.Seek(0)
			} else {
				i := player.CurrentIndex() - 1
				if i >= 0 {
					play(i, window)
				}
			}
		})
		back.Importance = widget.LowImportance
		back.Disable()

		next = widgets.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nil)
		next.Importance = widget.LowImportance
		next.OnTapped = func() {
			i := player.CurrentIndex() + 1
			q := player.Queue()
			if len(q) > i {
				play(i, window)
			}
		}
		next.Disable()

		settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			settings(window)
		})
		settingsButton.Importance = widget.LowImportance

		songProgressSlider = widget.NewSlider(0, 0)
		songProgressSlider.Disable()

		songVolumeSlider = widget.NewSlider(-2, 2)
		songVolumeSlider.Step = 0.01

		posd, fulld = widget.NewRichText(&widget.TextSegment{Text: "0:00", Style: widget.RichTextStyle{ColorName: theme.ColorNameDisabled}}), widget.NewRichText(&widget.TextSegment{Text: "-:--", Style: widget.RichTextStyle{ColorName: theme.ColorNameDisabled}})
		prevf := 0.0

		songProgressSlider.OnChanged = func(f float64) {
			passed := time.Duration(f) * time.Millisecond

			posd.Segments[0].(*widget.TextSegment).Text = durString(passed)
			posd.Segments[0].(*widget.TextSegment).Style.ColorName = theme.ColorNameForeground
			posd.Refresh()
			if math.Abs(f-prevf) > 1000 {
				player.Seek(time.Duration(f) * time.Millisecond)
				cs := player.CurrentSong()
				var in int
				for _, lyric := range cs.Lyrics.SyncedLyrics {
					if lyric.At <= passed {
						in++
					}
				}
				syncedLyrics = cs.Lyrics.SyncedLyrics[in:]
				lyricsTxt.SetCurrentLine(in + 1)
				lyricsTxt.Refresh()
			}
			prevf = f
		}

		volumeIcon := widget.NewIcon(theme.VolumeMuteIcon())

		songVolumeSlider.OnChanged = func(f float64) {
			player.SetMute(f == -1)
			player.SetVolume(f)
			switch {
			case f == -2:
				volumeIcon.SetResource(theme.VolumeMuteIcon())
			case f >= 0:
				volumeIcon.SetResource(theme.VolumeUpIcon())
			case f < 0:
				volumeIcon.SetResource(theme.VolumeDownIcon())
			}
		}

		progress := container.NewBorder(nil, nil, posd, fulld, songProgressSlider)
		control := container.NewGridWithRows(2, container.NewHBox(layout.NewSpacer(), back, pause, next, layout.NewSpacer()), progress)

		bottom = container.NewGridWithColumns(3, layout.NewSpacer(), control, container.NewGridWithColumns(2, layout.NewSpacer(),
			container.NewBorder(nil, nil, volumeIcon, nil, songVolumeSlider),
		))

		window.SetContent(container.NewStack(
			container.NewBorder(nil,
				container.NewStack(
					canvas.NewRectangle(color.Black),
					container.NewPadded(bottom),
				), nil, container.NewVBox(settingsButton), tabs),
			container.NewVBox(layout.NewSpacer(), widget.NewLabel(constant.APP_VERSION)),
		))

		size := window.Content().MinSize()
		size.Width *= 2.5
		size.Height *= 3
		window.Resize(size)

		window.Show()
		window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
			if ke.Name == "LeftShift" {
				if settingsDebugModeCheck != nil {
					if settingsDebugModeCheck.Hidden {
						settingsDebugModeCheck.Show()
					} else {
						settingsDebugModeCheck.Hide()
					}
				}
			}
		})

		tick := time.NewTicker(time.Millisecond)
		for {
			select {
			case <-tick.C:
				if !player.Playing() {
					continue
				}
				if songProgressSlider == nil {
					continue
				}
				passed := player.TimePosition()
				cs := player.CurrentSong()
				if preferences.Preferences.Bool("discord_rpc") {
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

				if len(syncedLyrics) == 0 {
					continue
				}
				lyric := syncedLyrics[0]
				if lyric.At <= passed {
					lyricsTxt.NextLine()
					syncedLyrics = syncedLyrics[1:]
					/*lyricsTxt.Segments[lyric.Index].(*widget.TextSegment).Style.TextStyle.Bold = true
					lyricsTxt.Refresh()
					if len(syncedLyrics) == 0 {
						continue
					}
					syncedLyrics = syncedLyrics[1:]

					dy := lyricsScroll.Offset.Y - (float32(lyric.Index) * theme.TextHeadingSize())

					lyricsScroll.Scrolled(&fyne.ScrollEvent{
						Scrolled: fyne.NewDelta(0, dy),
					})*/
				}
			case <-player.SongFinished:
				{
					prevf = 0
					q := player.Queue()
					i := player.CurrentIndex() + 1

					if len(q) <= i {
						// stop
						stopPlayer()
						continue
					}

					// switch to next song
					play(i, window)
					continue
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

func stopPlayer() {
	tabs.DisableIndex(2)
	pause.Disable()
	next.Disable()

	fulld.Segments[0].(*widget.TextSegment).Text = "-:--"
	fulld.Segments[0].(*widget.TextSegment).Style.ColorName = theme.ColorNameDisabled

	fulld.Refresh()

	posd.Segments[0].(*widget.TextSegment).Text = "0:00"
	posd.Segments[0].(*widget.TextSegment).Style.ColorName = theme.ColorNameDisabled

	posd.Refresh()

	songProgressSlider.Disable()

	bottom.Objects[0] = layout.NewSpacer()
	bottom.Refresh()
}
