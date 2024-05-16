package player

import (
	"qusic/lyrics"
	"qusic/streamer"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	yt "github.com/kkdai/youtube/v2"
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
	Video  *yt.Video
	Format *yt.Format

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
	queue []*Song

	paused      bool
	currentSong int

	streamer  *streamer.Streamer
	resampler *beep.Resampler

	SongFinished chan struct{}

	Downloader func(video *yt.Video) (beep.StreamSeekCloser, beep.Format, error)
}

func New(src Source) *Player {
	st := streamer.NewStreamer()
	p := &Player{
		queue:  make([]*Song, 0),
		Source: src,

		streamer:  st,
		resampler: beep.ResampleRatio(4, 1, st),

		currentSong: -1,

		SongFinished: make(chan struct{}),

		Downloader: streamer.New,
	}

	speaker.Init(streamer.SampleRate, streamer.SampleRate.N(time.Second/10))

	return p
}

func (p *Player) SetSpeed(x float64) {
	p.resampler.SetRatio(x)
}

func (p *Player) Volume() float64 {
	return p.streamer.Volume()
}

func (p *Player) SetVolume(v float64) {
	p.streamer.SetVolume(v)
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
	p.streamer.SetPaused(p.paused)
	return nil
}

func (p *Player) Paused() bool {
	return p.paused
}

func (p *Player) CurrentSong() *Song {
	return p.queue[p.currentSong]
}

func (p *Player) Playing() bool {
	return p.currentSong != -1
}

func (p *Player) postodur(pos int) time.Duration {
	return streamer.SampleRate.D(pos)
}

func (p *Player) durtopos(dur time.Duration) int {
	n := streamer.SampleRate.N(dur)
	return n
}

// Returns the current time position in the song
func (p *Player) TimePosition() time.Duration {
	return p.postodur(p.streamer.Position())
}

// Returns the time remaining for the song
func (p *Player) TimeRemaining(ms bool) time.Duration {
	pos := p.TimePosition()
	l := p.postodur(p.streamer.Len())

	return l - pos
}

func (p *Player) SeekRaw(pos int) error {
	return p.streamer.Seek(pos)
}

func (p *Player) Seek(dur time.Duration) error {
	return p.SeekRaw(p.durtopos(dur))
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
	streamer, _, err := p.Downloader(p.queue[i].Video)

	if err != nil {
		return err
	}

	p.currentSong = i

	speaker.Clear()

	p.streamer.SetStreamer(streamer)
	speaker.Play(beep.Seq(p.resampler, beep.Callback(func() {
		p.SongFinished <- struct{}{}
	})))

	return nil
}
