package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetSongLyricsOVH(artist, title string) (Song, error) {
	s := Song{LyricSource: "lyrics.ovh"}
	res, err := http.Get(fmt.Sprintf("https://api.lyrics.ovh/v1/%s/%s", url.PathEscape(artist), url.PathEscape(title)))
	if err != nil {
		return s, err
	}
	if res.StatusCode == 404 {
		return s, ErrNotFound
	}
	var data struct {
		Lyrics string `json:"lyrics"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return s, err
	}
	s.PlainLyrics = data.Lyrics
	s.PlainLyrics = s.PlainLyrics[strings.Index(s.PlainLyrics, "\n")+1:]
	for {
		i := strings.Index(s.PlainLyrics, "[")
		if i == -1 {
			break
		}
		i1 := strings.Index(s.PlainLyrics, "]")
		if s.PlainLyrics[i1+1] == '\n' {
			i1++
		}
		s.PlainLyrics = s.PlainLyrics[i-1 : i1+1]
	}
	return s, nil
}
