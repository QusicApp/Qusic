package main

import (
	"fmt"
	"image"
	"net/http"
	"qusic/logger"
	"qusic/lyrics"
	pl "qusic/player"
	"qusic/preferences"
	"qusic/widgets"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	fynesyncedlyrics "github.com/dweymouth/fyne-lyrics"
)

func getLyrics(song *pl.Song) {
	lyricsTxt = fynesyncedlyrics.NewLyricsViewer()
	lyricsTxt.TextSizeName = theme.SizeNameHeadingText
	lyricPage.Objects[2] = lyricsTxt
	lyricPage.Refresh()

	source := preferences.Preferences.String("lyrics.source")

	var err error
	switch source {
	case "lrclib":
		song.Lyrics, err = lyrics.GetSongLRCLIB(song.Name, song.Artists[0].Name, song.Album.Name, song.Duration, false)
	case "spotify":
		song.Lyrics, err = lyrics.GetLyricsSpotify(spotifyClient, song.ID)
	case "youtubesub":
		song.Lyrics, err = lyrics.GetLyricsYouTubeSubtitles(song.Video)
	case "ytmusicsynced":
		song.Lyrics, err = lyrics.GetSongYTMusicSynced(song.Video.ID)
	case "ytmusicplain":
		song.Lyrics, err = lyrics.GetSongYTMusicPlain(song.Video.ID)
	case "genius":
		song.Lyrics, err = lyrics.GetSongGenius(song.Artists[0].Name, song.Name)
	case "lyrics.ovh":
		song.Lyrics, err = lyrics.GetSongLyricsOVH(song.Artists[0].Name, song.Name)
	}
	if err != nil {
		logger.Errorf("No lyrics for %s<source:%s,album:%s,artist:%s,duration:%s>:%v", song.Name, source, song.Album.Name, song.Artists[0].Name, song.Duration, err)
	} else {
		logger.Infof("Fetched lyrics for %s (source: %s)", song.Name, source)
	}

	syncedLyrics = song.Lyrics.SyncedLyrics

	lyricsAlt.Hide()
	lyricsTxt.Show()

	if len(syncedLyrics) == 0 {
		if song.Lyrics.PlainLyrics == "" {
			lyricsAlt.ParseMarkdown("# Sorry, no lyrics were found for this song. Maybe try a different source.")
			lyricsAlt.Show()
			lyricsTxt.Hide()
		} else {
			lyricsTxt.SetLyrics(strings.Split(song.Lyrics.PlainLyrics, "\n"), false)
		}
	} else {
		var lines = make([]string, len(syncedLyrics))
		for i, lyric := range syncedLyrics {
			lines[i] = lyric.Lyric
		}
		lyricsTxt.SetLyrics(lines, true)
	}
	tabs.EnableIndex(2)
}

func setPlayedSong(song *pl.Song, w fyne.Window) {
	logger.Infof("Now played song: %s (%s)", song.Name, song.URL)

	getLyrics(song)

	pause.SetIcon(theme.MediaPauseIcon())

	d, err := http.Get(song.Thumbnails.Min().URL)
	if err != nil {
		return
	}
	image, _, _ := image.Decode(d.Body)
	img := canvas.NewImageFromImage(image)

	if preferences.Preferences.Bool("hardware_acceleration") {
		img.ScaleMode = canvas.ImageScaleFastest
	}

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

func play(i int, w fyne.Window) {
	q := player.Queue()
	s := q[i]
	player.GetVideo(s)

	logger.Inff("Playing song %s: ", s.Name)
	err := player.Play(i)
	songProgressSlider.SetValue(0)
	logger.Println(err)

	setPlayedSong(s, w)
	if err != nil {
		dialog.NewError(fmt.Errorf("There was an error playing your song, please check logs"), w).Show()
		stopPlayer()
	}
	if err != nil {
		return
	}
}

func playnow(so *pl.Song, w fyne.Window) {
	player.GetVideo(so)

	logger.Inff("Playing song %s: ", so.Name)
	err := player.PlayNow(so)
	songProgressSlider.SetValue(0)
	logger.Println(err)

	setPlayedSong(so, w)
	if err != nil {
		dialog.NewError(fmt.Errorf("There was an error playing your song, please check logs"), w).Show()
		stopPlayer()
	}
	if err != nil {
		return
	}
}
