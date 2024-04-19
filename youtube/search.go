package youtube

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kkdai/youtube/v2"
)

var Client = new(youtube.Client)

type SearchResult struct {
	ID, Title string
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
	for {
		i := strings.Index(s, `"videoId":"`)
		if i == -1 {
			break
		}
		s = s[i+len(`"videoId":"`):]

		id := s[:strings.Index(s, `"`)]

		i = strings.Index(s, `"title":{"runs":[{"text":"`)
		s = s[i+len(`"title":{"runs":[{"text":"`):]

		name := s[:strings.Index(s, `"`)]

		vids = append(vids, SearchResult{id, name})
	}
	return vids, nil
}
