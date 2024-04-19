package main

import (
	"fmt"
	"os"
	"qusic/spotify"
	"qusic/youtube"

	ytdl "github.com/kkdai/youtube/v2"

	"github.com/gen2brain/go-mpv"
)

var id = os.Getenv("SPOTIFY_ID")
var secret = os.Getenv("SPOTIFY_SECRET")

var client = spotify.New(id, secret)

var (
	currentSong spotify.TrackObject
	paused      bool
	songChanged = make(chan struct{}, 1)
)

func pauseResume() {
	player.Command([]string{"cycle", "pause"})
	paused = !paused
}

func seek(absSeconds int) {
	player.Command([]string{"seek", fmt.Sprint(absSeconds), "absolute"})
}

var player = mpv.New()

func init() {
	player.SetOptionString("audio-display", "no")
	player.SetOptionString("video", "no")
	player.SetOptionString("terminal", "no")
	player.SetOptionString("demuxer-max-bytes", "30MiB")
	player.SetOptionString("audio-client-name", "stmp")

	player.Initialize()
}

func playSong(song spotify.TrackObject) {
	v, _ := youtube.Search(song.Name + " - " + song.Artists[0].Name + " - " + song.Album.Name)
	vid := v[0]
	d, _ := vid.Data()
	var format ytdl.Format
	for _, f := range d.Formats.Type("audio") {
		if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			format = f
			break
		}
	}

	player.Command([]string{"loadfile", format.URL})
	currentSong = song
	songChanged <- struct{}{}
}
