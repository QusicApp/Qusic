package youtube

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type Named struct {
	ID, Name string
}

type Artist struct {
	Named
	Thumbnails  []Thumbnail
	Subscribers int
}

type Album struct {
	Named
	Artists    []Artist
	Year       int
	Thumbnails []Thumbnail
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Video struct {
	Thumbnails     []Thumbnail
	VideoID, Title string
	Album          Named
	Authors        []Named
	Duration       time.Duration

	Plays int
}

func (s *Video) Data() (*youtube.Video, error) {
	return Client.GetVideo(s.VideoID)
}

func jsonBody(d any) io.Reader {
	var buf = new(bytes.Buffer)

	json.NewEncoder(buf).Encode(d)

	return buf
}

type MusicSearchResult struct {
	TopResult Video

	Songs   []Video
	Videos  []Video
	Artists []Artist
	Albums  []Album
}

// YouTube Music API
type MusicClient struct{}

const (
	ParamSongsOnly  = "Eg-KAQwIARAAGAAgACgAMABqChAEEAMQCRAFEAo%3D"
	ParamVideosOnly = "Eg-KAQwIABABGAAgACgAMABqChAEEAMQCRAFEAo%3D"
)

func (m *MusicClient) Search(query string) (MusicSearchResult, error) {
	var result MusicSearchResult
	res, err := m.search(query, "")
	if err != nil {
		return result, err
	}
	contents := res.Contents.TabbedSearchResultsRenderer.Tabs.Title("YT Music").TabRenderer.Content.SectionListRenderer.Contents

	top := contents[0].MusicCardShelfRenderer
	result.TopResult = Video{
		Thumbnails: top.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
		VideoID:    top.OnTap.WatchEndpoint.VideoID,
		Title:      top.Title.Runs[0].Text,
		Album:      top.Subtitle.Album(),
		Authors:    top.Subtitle.Author(),
		Duration:   top.Subtitle.Duration(),

		Plays: -1,
	}

	songs := contents.Title("Songs")
	result.Songs = make([]Video, len(songs.MusicShelfRenderer.Contents))

	for i, song := range songs.MusicShelfRenderer.Contents {
		result.Songs[i] = Video{
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			VideoID:    song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			Album:      song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Album(),
			Authors:    song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author(),
			Duration:   song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration(),
		}

		multiplier := 1.0
		if len(song.MusicResponsiveListItemRenderer.FlexColumns) == 3 {
			plays := song.MusicResponsiveListItemRenderer.FlexColumns[2].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text
			plays = strings.TrimSuffix(plays, " plays")
			if len(plays) == 0 {
				continue
			}
			switch plays[len(plays)-1] {
			case 'B':
				multiplier = 10 ^ 9
				plays = plays[:len(plays)-1]
			case 'M':
				multiplier = 10 ^ 6
				plays = plays[:len(plays)-1]
			case 'K':
				multiplier = 10 ^ 3
				plays = plays[:len(plays)-1]
			}
			n, err := strconv.ParseFloat(plays, 64)
			if err != nil {
				continue
			}
			result.Songs[i].Plays = int(n * multiplier)
		}
	}

	if len(result.TopResult.Authors) == 0 {
		if len(result.Songs) != 0 {
			result.TopResult = result.Songs[0]
			result.Songs = result.Songs[1:]
		} else {
			result.TopResult = Video{}
		}
	}

	videos := contents.Title("Videos")
	result.Videos = make([]Video, len(videos.MusicShelfRenderer.Contents))

	for i, video := range videos.MusicShelfRenderer.Contents {
		result.Videos[i] = Video{
			Thumbnails: video.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			VideoID:    video.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Title:      video.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			Authors:    video.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author(),
			Duration:   video.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration(),
		}

		multiplier := 1.0
		if len(video.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs) == 7 {
			plays := video.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[4].Text
			plays = strings.TrimSuffix(plays, " views")
			if len(plays) == 0 {
				continue
			}
			switch plays[len(plays)-1] {
			case 'B':
				multiplier = 10 ^ 9
				plays = plays[:len(plays)-1]
			case 'M':
				multiplier = 10 ^ 6
				plays = plays[:len(plays)-1]
			case 'K':
				multiplier = 10 ^ 3
				plays = plays[:len(plays)-1]
			}
			n, err := strconv.ParseFloat(plays, 64)
			if err != nil {
				continue
			}
			result.Videos[i].Plays = int(n * multiplier)
		}
	}

	return result, nil
}

// TODO add views
func (m *MusicClient) SearchVideos(query string) ([]Video, error) {
	res, err := m.search(query, ParamVideosOnly)
	if err != nil {
		return nil, err
	}
	var results []Video

	for _, song := range res.Contents.TabbedSearchResultsRenderer.Tabs.Title("YT Music").TabRenderer.Content.SectionListRenderer.Contents.Title("Videos").MusicShelfRenderer.Contents {
		if len(song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs) != 5 {
			continue
		}

		authors := song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author()
		if len(authors) == 0 {
			authors = append(authors, Named{Name: song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text})
		}

		results = append(results, Video{
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			VideoID:    song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			Authors:    authors,
			Duration:   song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration(),
		})
	}

	return results, nil
}

func (m *MusicClient) SearchSongs(query string) ([]Video, error) {
	res, err := m.search(query, ParamSongsOnly)
	if err != nil {
		return nil, err
	}

	content := res.Contents.TabbedSearchResultsRenderer.Tabs.Title("YT Music").TabRenderer.Content.SectionListRenderer.Contents.Title("Songs")
	var results = make([]Video, len(content.MusicShelfRenderer.Contents))
	for i, song := range content.MusicShelfRenderer.Contents {
		results[i] = Video{
			Title:      song.MusicResponsiveListItemRenderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text,
			VideoID:    song.MusicResponsiveListItemRenderer.PlaylistItemData.VideoID,
			Thumbnails: song.MusicResponsiveListItemRenderer.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails,
			Authors:    song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Author(),
			Album:      song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Album(),
			Duration:   song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Duration(),
		}
		if len(results[i].Authors) == 0 {
			results[i].Authors = append(results[i].Authors, Named{Name: song.MusicResponsiveListItemRenderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text})
		}

		multiplier := 1.0

		if len(song.MusicResponsiveListItemRenderer.FlexColumns) == 3 {
			plays := song.MusicResponsiveListItemRenderer.FlexColumns[2].MusicResponsiveListItemFlexColumnRenderer.Text.Runs[0].Text
			plays = strings.TrimSuffix(plays, " plays")
			if len(plays) == 0 {
				continue
			}
			switch plays[len(plays)-1] {
			case 'B':
				multiplier = 10 ^ 9
				plays = plays[:len(plays)-1]
			case 'M':
				multiplier = 10 ^ 6
				plays = plays[:len(plays)-1]
			case 'K':
				multiplier = 10 ^ 3
				plays = plays[:len(plays)-1]
			}
			n, err := strconv.ParseFloat(plays, 64)
			if err != nil {
				continue
			}
			results[i].Plays = int(n * multiplier)
		}
	}

	return results, nil
}

func (*MusicClient) search(query, params string) (musicSearchResponse, error) {
	req, _ := http.NewRequest("POST", "https://music.youtube.com/youtubei/v1/search?prettyPrint=false", jsonBody(newMusicRequest(musicRequest{Query: query, Params: params})))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return musicSearchResponse{}, nil
	}
	defer res.Body.Close()

	var response musicSearchResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	return response, err
}

func (*MusicClient) LyricsBrowseID(videoId string) (string, error) {
	req, _ := http.NewRequest("POST", "https://music.youtube.com/youtubei/v1/next?prettyPrint=false", jsonBody(newMusicRequest(musicRequest{
		VideoID: videoId,
	})))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var response musicNextResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	tabs := response.Contents.SingleColumnMusicWatchNextResultsRenderer.TabbedRenderer.WatchNextTabbedResultsRenderer.Tabs
	for _, tab := range tabs {
		if tab.TabRenderer.Title == "Lyrics" {
			return tab.TabRenderer.Endpoint.BrowseEndpoint.BrowseID, nil
		}
	}
	return "", nil
}

func (c *MusicClient) Lyrics(videoId string) (Lyrics, error) {
	browse, err := c.LyricsBrowseID(videoId)
	if err != nil {
		return Lyrics{}, err
	}

	req, _ := http.NewRequest("POST", "https://music.youtube.com/youtubei/v1/browse?prettyPrint=false", jsonBody(newMusicRequest(musicRequest{
		BrowseID: browse,
	})))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Lyrics{}, err
	}
	defer res.Body.Close()

	var response musicBrowseLyricsResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return Lyrics{}, err
	}

	return Lyrics{
		Source: strings.TrimPrefix(response.Contents.SectionListRenderer.Contents[0].MusicDescriptionShelfRenderer.Footer.Runs[0].Text, "Source: "),
		Lyrics: response.Contents.SectionListRenderer.Contents[0].MusicDescriptionShelfRenderer.Description.Runs[0].Text,
	}, nil
}

// WIP
func (*MusicClient) PlaybackData(videoId string) (musicPlayerResponse, error) {
	req, _ := http.NewRequest("POST", "https://music.youtube.com/youtubei/v1/player?prettyPrint=false", jsonBody(newMusicRequest(musicRequest{
		VideoID: videoId,
	})))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return musicPlayerResponse{}, nil
	}
	defer res.Body.Close()

	var response musicPlayerResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	return response, err
}
