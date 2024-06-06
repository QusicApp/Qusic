package lyrics

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type songData struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func parseSyncedLyrics(str string) []SyncedLyric {
	lines := strings.Split(str, "\n")
	syncedLyrics := make([]SyncedLyric, len(lines))
	for index, line := range lines {
		i := strings.Index(line, " ")
		if i == -1 {
			continue
		}
		stamp := line[:i]
		lyric := strings.TrimSpace(line[i:])

		stamp = stamp[1 : len(stamp)-1]
		sep := strings.Split(stamp, ":")
		if len(sep) != 2 {
			continue
		}
		minutes, err := strconv.ParseInt(sep[0], 10, 64)
		if err != nil {
			continue
		}
		seconds, err := strconv.ParseFloat(sep[1], 64)
		if err != nil {
			continue
		}
		if lyric == "" {
			lyric = "♪"
		}

		duration := (time.Duration(minutes) * time.Minute) + time.Duration(seconds*float64(time.Second))
		syncedLyrics[index] = SyncedLyric{
			At:    duration,
			Lyric: lyric,
			Index: index,
		}
	}
	return syncedLyrics
}

func SearchSongLRCLIB(trackName, artistName, albumName string) (Song, error) {
	s := Song{LyricSource: "LRCLIB"}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://lrclib.net/api/search?track_name=%s&artist_name=%s&album_name=%s",
		url.QueryEscape(trackName),
		url.QueryEscape(artistName),
		url.QueryEscape(albumName),
	), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return s, err
	}
	if res.StatusCode == http.StatusNotFound {
		return s, ErrNotFound
	}
	var data []songData
	err = json.NewDecoder(res.Body).Decode(&data)
	if len(data) == 0 {
		return s, ErrNotFound
	}
	s.PlainLyrics = data[0].PlainLyrics
	s.SyncedLyrics = parseSyncedLyrics(data[0].SyncedLyrics)

	return s, err
}

func GetSongLRCLIB(trackName, artistName, albumName string, duration time.Duration, cached bool) (Song, error) {
	s := Song{LyricSource: "LRCLIB"}
	c := ""
	if cached {
		c = "-cached"
	}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://lrclib.net/api/get%s?track_name=%s&artist_name=%s&album_name=%s&duration=%d",
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
	var data songData
	err = json.NewDecoder(res.Body).Decode(&data)
	s.PlainLyrics = data.PlainLyrics
	s.SyncedLyrics = parseSyncedLyrics(data.SyncedLyrics)

	return s, err
}

func PublishSongLRCLIB(
	trackName, artistName, albumName string, duration time.Duration, plainLyrics string, syncedLyrics []SyncedLyric, token string,
) error {
	var data = map[string]any{
		"trackName":    trackName,
		"artistName":   artistName,
		"albumName":    albumName,
		"duration":     duration / time.Second,
		"plainLyrics":  plainLyrics,
		"syncedLyrics": "",
	}
	minutes := int(duration.Minutes())
	seconds := duration.Seconds()
	for i, lyric := range syncedLyrics {
		newline := ""
		if i != len(syncedLyrics)-1 {
			newline = "\n"
		}
		data["syncedLyrics"] = data["syncedLyrics"].(string) + fmt.Sprintf("[%d:%02.2f] %s", minutes, seconds, lyric.Lyric) + newline
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if token == "" {
		token, err = NewLRCLIBPublishToken()
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodPost, "https://lrclib.net/api/publish", bytes.NewReader(body))
	req.Header.Set("X-Publish-Token", token)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusCreated {
		return nil
	}
	var failed PublishError
	err = json.NewDecoder(res.Body).Decode(&failed)
	if err != nil {
		return err
	}
	return failed
}

type PublishError struct {
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (p PublishError) Error() string {
	return fmt.Sprintf("%s (code %d): %s", p.Name, p.Code, p.Message)
}

func NewLRCLIBPublishToken() (string, error) {
	req, err := http.NewRequest(http.MethodPost, "https://lrclib.net/api/request-challenge", nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var response struct {
		Prefix string `json:"prefix"`
		Target string `json:"target"`
	}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	nonce := solveChallenge(response.Prefix, response.Target)

	return fmt.Sprintf("{%s}:{%s}", response.Prefix, nonce), nil
}

func verifyNonce(result []byte, target []byte) bool {
	if len(result) != len(target) {
		return false
	}

	for i := 0; i < len(result); i++ {
		if result[i] > target[i] {
			return false
		} else if result[i] < target[i] {
			break
		}
	}

	return true
}

func solveChallenge(prefix string, targetHex string) string {
	nonce := 0
	target, _ := hex.DecodeString(targetHex)

	for {
		input := prefix + strconv.Itoa(nonce)
		hashed := sha256.Sum256([]byte(input))

		if verifyNonce(hashed[:], target) {
			break
		} else {
			nonce++
		}
	}

	return strconv.Itoa(nonce)
}
