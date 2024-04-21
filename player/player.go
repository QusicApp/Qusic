package player

import (
	"errors"
	"fmt"
	"qusic/spotify"
	"qusic/youtube"
	"strconv"
	"time"

	"github.com/gen2brain/go-mpv"
)

var ErrNotFoundSong = errors.New("not found song")

type Song struct {
	spotify.TrackObject
	URL string
}

type Player struct {
	client *spotify.Client
	player *mpv.Mpv
	queue  []*Song

	paused      bool
	currentSong int

	NowPlaying chan *Song
}

func New(client *spotify.Client) *Player {
	return &Player{
		client: client,
		player: mpv.New(),
		queue:  make([]*Song, 0),

		NowPlaying: make(chan *Song),
	}
}

func (p *Player) Initialize() {
	p.player.SetOptionString("audio-display", "no")
	p.player.SetOptionString("video", "no")
	p.player.SetOptionString("terminal", "no")
	p.player.SetOptionString("demuxer-max-bytes", "30MiB")
	p.player.SetOptionString("audio-client-name", "stmp")

	p.player.Initialize()
}

func (p *Player) Song(song spotify.TrackObject) *Song {
	v, _ := youtube.Search(song.Name + " - " + song.Artists[0].Name + " - " + song.Album.Name)
	vid := v[0]
	d, _ := vid.Data()
	format := d.Formats.Type("audio")[0]

	return &Song{song, format.URL}
}

func (p *Player) AddToQueue(song *Song) {
	p.queue = append(p.queue, song)
}

func (p *Player) PauseCycle() error {
	p.paused = !p.paused
	return p.player.Command([]string{"cycle", "pause"})
}

func (p *Player) Paused() bool {
	return p.paused
}

func (p *Player) CurrentSong() *Song {
	return p.queue[p.currentSong]
}

func (p *Player) Playing() bool {
	return p.player.GetPropertyString("filename") != ""
}

// Returns the time position in seconds (or in milliseconds in if ms is true)
func (p *Player) TimePositionRaw(ms bool) (float64, error) {
	cmd := "time-pos"
	if ms {
		cmd += "/full"
	}
	return strconv.ParseFloat(p.player.GetPropertyString(cmd), 64)
}

// Returns the time remaining in seconds (or in milliseconds in if ms is true)
func (p *Player) TimeRemainingRaw(ms bool) (float64, error) {
	cmd := "time-remaining"
	if ms {
		cmd += "/full"
	}
	return strconv.ParseFloat(p.player.GetPropertyString(cmd), 64)
}

func (p *Player) TimePosition(ms bool) (time.Duration, error) {
	d, err := p.TimePositionRaw(ms)
	if err != nil {
		return 0, err
	}
	dur := time.Duration(d)
	if ms {
		return dur * time.Millisecond, nil
	} else {
		return dur * time.Second, nil
	}
}

func (p *Player) TimeRemaining(ms bool) (time.Duration, error) {
	d, err := p.TimeRemainingRaw(ms)
	if err != nil {
		return 0, err
	}
	dur := time.Duration(d)
	if ms {
		return dur * time.Millisecond, nil
	} else {
		return dur * time.Second, nil
	}
}

func (p *Player) Seek(absSeconds int) error {
	return p.player.Command([]string{"seek", fmt.Sprint(absSeconds), "absolute"})
}

func (p *Player) ClearQueue() {
	clear(p.queue)
}

func (p *Player) PlayNow(s *Song) error {
	p.ClearQueue()
	p.AddToQueue(s)
	if p.paused {
		if err := p.PauseCycle(); err != nil {
			return err
		}
	}
	return p.Play(0)
}

func (p *Player) Play(i int) error {
	if i < 0 || len(p.queue) < i {
		return nil
	}

	if err := p.player.Command([]string{"loadfile", p.queue[i].URL}); err != nil {
		return err
	}

	p.NowPlaying <- p.queue[i]
	p.currentSong = i
	return nil
}
