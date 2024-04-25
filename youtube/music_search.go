package youtube

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"qusic/lyrics"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type MusicSearchResult struct {
	Thumbnails []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	ID, Title, Author, Album string
	Duration                 time.Duration

	SongInfo lyrics.Song `json:"-"`
}

func (s *MusicSearchResult) Data() (*youtube.Video, error) {
	return Client.GetVideo(s.ID)
}

func (o *MusicSearchResult) FetchSongInfo() (err error) {
	o.SongInfo, err = lyrics.SearchSongLRCLIB(o.Title, o.Author, o.Album)
	return
}

type musicSearchRequest struct {
	Context struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
		} `json:"client"`
	} `json:"context"`
	Params string `json:"params"`
	Query  string `json:"query"`
}

func newMusicSearchRequest(query, params string) musicSearchRequest {
	d := musicSearchRequest{Query: query, Params: params}
	d.Context.Client.ClientName = "WEB_REMIX"
	d.Context.Client.ClientVersion = "1.20240417.01.01"

	return d
}

type musicThumbnailRenderer struct {
	Thumbnail struct {
		Thumbnails []struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnails"`
	} `json:"thumbnail"`
}

type runs struct {
	Runs []struct {
		Text               string `json:"text"`
		NavigationEndpoint struct {
			WatchEndpoint struct {
				VideoID string `json:"videoId"`
			} `json:"watchEndpoint"`
		} `json:"navigationEndpoint"`
	} `json:"runs"`
}

type musicSearchResponse struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
								MusicCardShelfRenderer struct {
									Thumbnail struct {
										MusicThumbnailRenderer musicThumbnailRenderer `json:"musicThumbnailRenderer"`
									} `json:"thumbnail"`

									Title runs `json:"title"`
								} `json:"musicCardShelfRenderer"`
								MusicShelfRenderer struct {
									Title    runs `json:"title"`
									Contents []struct {
										MusicResponsiveListItemRenderer struct {
											PlaylistItemData struct {
												VideoID string `json:"videoId"`
											} `json:"playlistItemData"`
											FlexColumns []struct {
												MusicResponsiveListItemFlexColumnRenderer struct {
													Text runs `json:"text"`
												} `json:"musicResponsiveListItemFlexColumnRenderer"`
											} `json:"flexColumns"`
											Thumbnail struct {
												MusicThumbnailRenderer musicThumbnailRenderer `json:"musicThumbnailRenderer"`
											} `json:"thumbnail"`
										} `json:"musicResponsiveListItemRenderer"`
									} `json:"contents"`
								} `json:"musicShelfRenderer"`
							} `json:"contents"`
						} `json:"sectionListRenderer"`
					} `json:"content"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"tabbedSearchResultsRenderer"`
	} `json:"contents"`
}

func jsonBody(d any) io.Reader {
	var buf = new(bytes.Buffer)

	json.NewEncoder(buf).Encode(d)

	return buf
}

type MusicClient struct {
}

func (m MusicClient) SearchSongs(query string) ([]MusicSearchResult, error) {
	res, err := m.search(query, "Eg-KAQwIARAAGAAgACgAMABqChAEEAMQCRAFEAo%3D")
	if err != nil {
		return nil, err
	}
	var results = make([]MusicSearchResult, len(res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents[0].MusicShelfRenderer.Contents))

	for i, song := range res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents[0].MusicShelfRenderer.Contents {
		results[i] = MusicSearchResult{
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			ID:         song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			Author:     song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			Album:      song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[2].Text,
		}
		dur := song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[4].Text

		sep := strings.Split(dur, ":")
		var (
			hours, minutes, seconds int64
		)
		switch len(sep) {
		case 3:
			hours, _ = strconv.ParseInt(sep[0], 10, 64)
			minutes, _ = strconv.ParseInt(sep[1], 10, 64)
			seconds, _ = strconv.ParseInt(sep[2], 10, 64)
		case 2:
			minutes, _ = strconv.ParseInt(sep[0], 10, 64)
			seconds, _ = strconv.ParseInt(sep[1], 10, 64)
		}

		duration := (time.Duration(hours) * time.Hour) + (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second)

		results[i].Duration = duration
	}

	return results, nil
}

func (MusicClient) search(query, params string) (musicSearchResponse, error) {
	req, _ := http.NewRequest("POST", "https://music.youtube.com/youtubei/v1/search?prettyPrint=false", jsonBody(newMusicSearchRequest(query, params)))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return musicSearchResponse{}, nil
	}
	defer res.Body.Close()

	var response musicSearchResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	return response, err
}
