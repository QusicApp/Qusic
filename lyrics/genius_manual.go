package lyrics

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Didn't feel like using API

type SyncedLyric struct {
	At    time.Duration
	Lyric string
	Index int
}

type Song struct {
	Description  string
	PlainLyrics  string
	SyncedLyrics []SyncedLyric
}

var ErrNotFound = errors.New("not found song")

func GetSongGeniusManually(artistName, trackName string) (Song, error) {
	s := Song{}
	url := "https://genius.com/" + strings.ReplaceAll(artistName, " ", "-") + "-" + strings.ReplaceAll(trackName, " ", "-") + "-lyrics"
	res, err := http.Get(url)
	if err != nil {
		return s, err
	}
	if res.StatusCode == http.StatusNotFound {
		return s, ErrNotFound
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return s, err
	}
	str := string(b)

	for {
		i := strings.Index(str, "<meta content=\"")
		if i == -1 {
			break
		}
		in := i + len("<meta content=\"")
		i1 := strings.IndexRune(str[in:], '"')
		data := str[in:][:i1]

		isDescription := strings.HasPrefix(str[in+i1:], `" property="og:description"`)
		if isDescription {
			s.Description = data
			break
		}

		str = str[in+i1:]
	}

	lines := strings.Split(str, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "      window.__PRELOADED_STATE__ =") {
			l = strings.TrimPrefix(l, "      window.__PRELOADED_STATE__ = JSON.parse('")
			l = strings.TrimSuffix(l, "');")
			l = strings.ReplaceAll(l, "\\", "")
			l = strings.ReplaceAll(l, "<br>", "\\n")

			i0 := strings.Index(l, "[Verse 1]")
			if i0 != -1 {
				l = l[i0:]
			}

			l := l
			for {
				i := strings.Index(l, `data-id="`)
				if i == -1 {
					break
				}
				l = l[i+len(`data-id="`):]
				i1 := strings.Index(l, "<")
				i2 := strings.Index(l[:i1], ">")
				lyric := l[:i1][i2+1:]
				if lyric[0] == '[' {
					continue
				}
				lyric = strings.ReplaceAll(lyric, "\\nn", "\n")
				s.PlainLyrics += lyric + "\n"
				c := l[i2+1+len(lyric):][:20]
				if strings.Contains(c, "V") || strings.Contains(c, "C") {
					s.PlainLyrics += "\n"
				}
			}
			break
		}
	}
	s.PlainLyrics = s.PlainLyrics[:len(s.PlainLyrics)-1]

	return s, nil
}
