package lyrics

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type YouTubeSubtitlesResponse struct {
	XMLName  xml.Name `xml:"transcript"`
	TextList []struct {
		Start string `xml:"start,attr"`
		Dur   string `xml:"dur,attr"`
		Value string `xml:",chardata"`
	} `xml:"text"`
}

func GetLyricsYouTubeSubtitles(video *youtube.Video) (Song, error) {
	s := Song{LyricSource: "YouTube Subtitles"}
	if len(video.CaptionTracks) == 0 {
		return s, ErrNotFound
	}
	res, err := http.Get(strings.TrimSuffix(video.CaptionTracks[0].BaseURL, "&fmt=srv3"))
	if err != nil {
		return s, err
	}
	var subtitles YouTubeSubtitlesResponse
	err = xml.NewDecoder(res.Body).Decode(&subtitles)
	if err != nil {
		return s, err
	}
	var index int
	for _, subtitle := range subtitles.TextList {
		if subtitle.Value[0] == '(' || subtitle.Value[0] == '[' {
			continue
		}
		subtitle.Value = strings.ReplaceAll(subtitle.Value, "&#39;", "'")
		subtitle.Value = strings.ReplaceAll(subtitle.Value, "&quot;", `"`)
		subtitle.Value = strings.ReplaceAll(subtitle.Value, "♪", "")

		secs, _ := strconv.ParseFloat(subtitle.Start, 64)
		s.SyncedLyrics = append(s.SyncedLyrics, SyncedLyric{
			At:    time.Duration(secs) * time.Second,
			Lyric: subtitle.Value,
			Index: index,
		})
		index++
	}
	return s, nil
}
