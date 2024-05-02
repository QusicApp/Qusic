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

type Author struct {
	ID, Name string
}

type MusicSearchResult struct {
	Thumbnails []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	VideoID, Title string
	Album          Author
	Authors        []Author
	Duration       time.Duration

	PlaysViews int

	SongInfo lyrics.Song `json:"-"`
}

func (s *MusicSearchResult) Data() (*youtube.Video, error) {
	return Client.GetVideo(s.VideoID)
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
			BrowseEndpoint struct {
				BrowseID                              string `json:"browseId"`
				BrowseEndpointContextSupportedConfigs struct {
					BrowseEndpointContextMusicConfig struct {
						PageType string `json:"pageType"`
					} `json:"browseEndpointContextMusicConfig"`
				} `json:"browseEndpointContextSupportedConfigs"`
			} `json:"browseEndpoint"`
			WatchEndpoint struct {
				VideoID string `json:"videoId"`
			} `json:"watchEndpoint"`
		} `json:"navigationEndpoint"`
	} `json:"runs"`
}

func (r runs) Author() []Author {
	authors := make([]Author, 0)
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_ARTIST" || run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_USER_CHANNEL" {
			authors = append(authors, Author{Name: run.Text, ID: run.NavigationEndpoint.BrowseEndpoint.BrowseID})
		}
	}
	return authors
}

func (r runs) Album() Author {
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_ALBUM" {
			return Author{Name: run.Text, ID: run.NavigationEndpoint.BrowseEndpoint.BrowseID}
		}
	}
	return Author{}
}

func (r runs) Duration() string {
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseID == "" && strings.Index(run.Text, ":") != -1 {
			return run.Text
		}
	}
	return ""
}

type musicSearchResponse struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
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

type MusicClient struct{}

const (
	ParamSongsOnly  = "Eg-KAQwIARAAGAAgACgAMABqChAEEAMQCRAFEAo%3D"
	ParamVideosOnly = "Eg-KAQwIABABGAAgACgAMABqChAEEAMQCRAFEAo%3D"
)

func (m *MusicClient) SearchVideos(query string) ([]MusicSearchResult, error) {
	res, err := m.search(query, ParamVideosOnly)
	if err != nil {
		return nil, err
	}
	var results []MusicSearchResult

	i := len(res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents) - 1

	for _, song := range res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents[i].MusicShelfRenderer.Contents {
		if len(song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs) != 5 {
			continue
		}
		dur := song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration()

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

		results = append(results, MusicSearchResult{
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			VideoID:    song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			Authors:    song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author(),
			Duration:   duration,
		})
	}

	return results, nil
}

func (m *MusicClient) SearchSongs(query string) ([]MusicSearchResult, error) {
	res, err := m.search(query, ParamSongsOnly)
	if err != nil {
		return nil, err
	}

	i := len(res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents) - 1

	var results = make([]MusicSearchResult, len(res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents[i].MusicShelfRenderer.Contents))
	for i, song := range res.Contents.TabbedSearchResultsRenderer.Tabs[0].TabRenderer.Content.SectionListRenderer.Contents[i].MusicShelfRenderer.Contents {
		dur := song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration()

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

		results[i] = MusicSearchResult{
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			VideoID:    song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			Authors:    song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author(),
			Album:      song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Album(),
			Duration:   duration,
		}
	}

	return results, nil
}

func (*MusicClient) search(query, params string) (musicSearchResponse, error) {
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
