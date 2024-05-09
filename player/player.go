package player

import (
	"fmt"
	"qusic/lyrics"
	"strconv"
	"time"

	yt "github.com/kkdai/youtube/v2"

	"github.com/gen2brain/go-mpv"
)

type Named struct {
	Name, URL, ID string
}

type Artist struct {
	Named
	Thumbnails Thumbnails
	Listeners  int
}

type Album struct {
	Named
	Thumbnails Thumbnails
	Year       int
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type SearchResult struct {
	TopResult Song

	Songs   []Song
	Artists []Artist
	Albums  []Album
}

type Thumbnails []Thumbnail

func (t Thumbnails) Min() Thumbnail {
	i := -1
	x, y := 0, 0
	for in, thumbnail := range t {
		if in == 0 {
			x, y = thumbnail.Width, thumbnail.Height
			i = 0
		}
		if thumbnail.Width < x && thumbnail.Height < y {
			x, y = thumbnail.Width, thumbnail.Height
			i = in
		}
	}
	if i == -1 {
		return Thumbnail{}
	}
	return t[i]
}

func (t Thumbnails) Max() Thumbnail {
	i := -1
	x, y := 0, 0
	for in, thumbnail := range t {
		if thumbnail.Width > x && thumbnail.Height > y {
			x, y = thumbnail.Width, thumbnail.Height
			i = in
		}
	}
	if i == -1 {
		return Thumbnail{}
	}
	return t[i]
}

const (
	YTMusic = iota
	Spotify
)

type Song struct {
	Video     *yt.Video
	StreamURL string

	Provider int

	Lyrics lyrics.Song

	Name, URL, ID string
	Album         Album
	Artists       []Artist
	Thumbnails    Thumbnails

	Duration time.Duration
	Plays    int
}

type Source interface {
	GetVideo(*Song)
	Search(query string) SearchResult
}

type Player struct {
	Source
	player *mpv.Mpv
	queue  []*Song

	paused      bool
	currentSong int
}

func New(src Source) *Player {
	return &Player{
		player: mpv.New(),
		queue:  make([]*Song, 0),
		Source: src,

		currentSong: -1,
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

func (p *Player) Volume() (float64, error) {
	return strconv.ParseFloat(p.player.GetPropertyString("ao-volume"), 64)
}

func (p *Player) SetVolume(v float64) error {
	return p.player.SetPropertyString("ao-volume", fmt.Sprint(v))
}

func (p *Player) Queue() []*Song {
	return p.queue
}

func (p *Player) CurrentIndex() int {
	return p.currentSong
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
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

// Returns the current time position in the song
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

// Returns the time remaining for the song
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

func (p *Player) SeekRaw(absSeconds int) error {
	return p.player.Command([]string{"seek", fmt.Sprint(absSeconds), "absolute"})
}

func (p *Player) Seek(dur time.Duration) error {
	return p.SeekRaw(int(dur / time.Second))
}

func (p *Player) ClearQueue() {
	p.queue = p.queue[:0]
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

func (p *Player) SetCurrentSong(i int) {
	p.currentSong = i
}

func (p *Player) Play(i int) error {
	if i < 0 || len(p.queue) < i {
		return nil
	}

	if err := p.player.Command([]string{"loadfile", p.queue[i].StreamURL}); err != nil {
		return err
	}

	p.currentSong = i
	return nil
}
