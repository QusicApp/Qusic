package youtube

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/kkdai/youtube/v2"
)

var Client = new(youtube.Client)

type SearchResult struct {
	ID, Title string
	Duration  time.Duration
}

func (s SearchResult) Data() (*youtube.Video, error) {
	return Client.GetVideo(s.ID)
}

func Search(q string) ([]SearchResult, error) {
	res, err := http.Get("https://www.youtube.com/results?search_query=" + url.QueryEscape(q))
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	s := string(b)

	s = s[strings.Index(s, `{"contents":[{"videoRenderer":`):]
	var vids []SearchResult
	for in := 0; ; in++ {
		i := strings.Index(s, `"videoId":"`)
		if i == -1 {
			break
		}
		s = s[i+len(`"videoId":"`):]

		id := s[:strings.Index(s, `"`)]

		for in > 0 && id == vids[in-1].ID {
			i := strings.Index(s, `"videoId":"`)
			if i == -1 {
				break
			}

			s = s[i+len(`"videoId":"`):]

			id = s[:strings.Index(s, `"`)]
		}

		i = strings.Index(s, `"title":{"runs":[{"text":"`)
		s = s[i+len(`"title":{"runs":[{"text":"`):]

		name := s[:strings.Index(s, `"`)]

		i = strings.Index(s, `"lengthText"`)
		s = s[i+len(`"lengthText`):]

		i = strings.Index(s, `"simpleText":"`)
		s = s[i+len(`"simpleText":"`):]

		duration := s[:strings.Index(s, `"`)]

		sp := strings.Split(duration, ":")
		if len(sp) == 0 {
			continue
		}
		var (
			hour, minute, second int64
		)
		switch len(sp) {
		case 3:
			hour, err = strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				continue
			}
			minute, err = strconv.ParseInt(sp[1], 10, 64)
			if err != nil {
				continue
			}
			second, err = strconv.ParseInt(sp[2], 10, 64)
			if err != nil {
				continue
			}
		case 2:
			minute, err = strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				continue
			}
			second, err = strconv.ParseInt(sp[1], 10, 64)
			if err != nil {
				continue
			}
		default:
			continue
		}

		h, m, s := *(*time.Duration)(unsafe.Pointer(&hour)), *(*time.Duration)(unsafe.Pointer(&minute)), *(*time.Duration)(unsafe.Pointer(&second))

		vids = append(vids, SearchResult{id, name, h*time.Hour + m*time.Minute + s*time.Second})
	}
	return vids, nil
}
