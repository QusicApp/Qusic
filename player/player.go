package player

import (
	"fmt"
	"qusic/lyrics"
	"qusic/spotify"
	"qusic/youtube"
	"strconv"
	"time"
	"unsafe"

	yt "github.com/kkdai/youtube/v2"

	"github.com/gen2brain/go-mpv"
)

type Artist struct {
	Name, URL, ID string
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Song struct {
	Video     *yt.Video
	StreamURL string

	Lyrics lyrics.Song

	Name, URL  string
	Album      Artist
	Artists    []Artist
	Thumbnails []Thumbnail

	Duration time.Duration
	Plays    int
}

func (o *Song) FetchSongInfo() (err error) {
	o.Lyrics, err = lyrics.GetSongLRCLIB(o.Name, o.Artists[0].Name, o.Album.Name, o.Duration, false)
	return
}

type Player struct {
	player *mpv.Mpv
	queue  []*Song

	paused      bool
	currentSong int
}

func New() *Player {
	return &Player{
		player: mpv.New(),
		queue:  make([]*Song, 0),
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

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func (p *Player) SpotifySong(s spotify.TrackObject) Song {
	v, _ := (*youtube.MusicClient).SearchSongs(nil, s.Artists[0].Name+" - "+s.Name)

	dur := *(*time.Duration)(unsafe.Pointer(&s.DurationMS)) * time.Millisecond
	var vid *youtube.MusicSearchResult
	for _, video := range v {
		if abs(video.Duration-dur) <= 2*time.Second {
			vid = &video
			break
		}
	}

	if vid == nil {
		vid = &v[0]
	}

	d, _ := vid.Data()
	format := d.Formats.Type("audio")[0]

	var artists = make([]Artist, len(s.Artists))
	for i, artist := range s.Artists {
		artists[i] = Artist{
			Name: artist.Name,
			ID:   artist.ID,
			URL:  artist.ExternalURLs.Spotify,
		}
	}

	return Song{
		Video: d, StreamURL: format.URL,

		Name:    s.Name,
		Artists: artists,
		Album: Artist{
			Name: s.Album.Name,
			ID:   s.Album.ID,
			URL:  s.Album.ExternalURLs.Spotify,
		},

		Duration: dur,
		URL:      s.ExternalURLs.Spotify,

		Thumbnails: *(*[]Thumbnail)(unsafe.Pointer(&s.Album.Images)),
	}
}

func (p *Player) YoutubeMusicSong(s youtube.MusicSearchResult) *Song {
	d, _ := s.Data()
	format := d.Formats.Type("audio")[0]

	var artists = make([]Artist, len(s.Authors))
	for i, artist := range s.Authors {
		artists[i] = Artist{
			Name: artist.Name,
			ID:   artist.ID,
			URL:  fmt.Sprintf("https://music.youtube.com/channel/%s", artist.ID),
		}
	}

	return &Song{
		Video: d, StreamURL: format.URL,

		Name:    s.Title,
		URL:     fmt.Sprintf("https://youtu.be/%s", s.VideoID),
		Artists: artists,
		Album: Artist{
			Name: s.Album.Name,
			ID:   s.Album.ID,
			URL:  fmt.Sprintf("https://music.youtube.com/browse/%s", s.Album.ID),
		},
		Thumbnails: *(*[]Thumbnail)(unsafe.Pointer(&s.Thumbnails)),

		Duration: s.Duration,
		Plays:    s.PlaysViews,
	}
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
