package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func GetSongLRCLIB(trackName, artistName, albumName string, duration time.Duration, cached bool) (Song, error) {
	s := Song{}
	c := ""
	if cached {
		c = "-cached"
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://lrclib.net/api/get%s?track_name=%s&artist_name=%s&album_name=%s&duration=%d",
		c,
		url.QueryEscape(trackName),
		url.QueryEscape(artistName),
		url.QueryEscape(albumName),
		duration/time.Second,
	), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return s, err
	}
	if res.StatusCode == http.StatusNotFound {
		return s, ErrNotFound
	}
	var data struct {
		ID           int           `json:"id"`
		TrackName    string        `json:"trackName"`
		ArtistName   string        `json:"artistName"`
		AlbumName    string        `json:"albumName"`
		Duration     time.Duration `json:"duration"`
		Instrumental bool          `json:"instrumental"`
		PlainLyrics  string        `json:"plainLyrics"`
		SyncedLyrics string        `json:"syncedLyrics"`
	}
	err = json.NewDecoder(res.Body).Decode(&data)
	s.PlainLyrics = data.PlainLyrics

	lines := strings.Split(data.SyncedLyrics, "\n")
	for _, line := range lines {
		i := strings.Index(line, " ")
		if i == -1 {
			continue
		}
		stamp := line[:i]
		lyric := line[i:]

		stamp = stamp[1 : len(stamp)-1]
		sep := strings.Split(stamp, ":")
		if len(sep) != 2 {
			continue
		}
		minutes, err := strconv.ParseInt(sep[0], 10, 64)
		if err != nil {
			continue
		}
		seconds, err := strconv.ParseFloat(sep[0], 64)
		if err != nil {
			continue
		}

		duration := (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second) + time.Duration(seconds*float64(time.Second))
		s.SyncedLyrics = append(s.SyncedLyrics, SyncedLyric{
			At:    duration,
			Lyric: lyric,
		})
	}

	return s, err
}
