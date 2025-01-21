package player

import (
	"github.com/qusicapp/qusic/preferences"
	"github.com/qusicapp/qusic/spotify"
	"github.com/qusicapp/qusic/youtube"
	"strconv"
	"strings"
	"time"
	"unsafe"

	yt "github.com/kkdai/youtube/v2"
)

var ytclient yt.Client

func NewSpotifySource(client *spotify.Client) SpotifySource {
	return SpotifySource{client: client}
}

type SpotifySource struct {
	client *spotify.Client
}

func (s SpotifySource) Client() *spotify.Client {
	return s.client
}

func (source SpotifySource) GetVideo(s *Song) {
	if preferences.Preferences.Bool("spotify.download_yt") || source.client.Cookie_sp_dc == "" {
		v, _ := (*youtube.MusicClient).SearchSongs(nil, s.Artists[0].Name+" - "+s.Name)

		var vid *youtube.Video
		for _, video := range v {
			if abs(video.Duration-s.Duration) <= 2*time.Second && strings.Contains(s.Name, video.Title) {
				vid = &video
				break
			}
		}

		if vid == nil {
			vid = &v[0]
		}

		video, err := ytclient.GetVideo(vid.VideoID)
		if err != nil {
			return
		}
		s.Video = video
	}
}

func (source SpotifySource) Search(query string) SearchResult {
	var result SearchResult
	res, _ := source.client.Search(query, spotify.QueryAll, "", nil, nil, false)
	if len(res.Tracks.Items) == 0 {
		goto csongs
	}
	result.TopResult = source.Song(res.Tracks.Items[0])

	result.Songs = make([]Song, len(res.Tracks.Items))
	for i, song := range res.Tracks.Items {
		result.Songs[i] = source.Song(song)
	}

csongs:
	result.Artists = make([]Artist, len(res.Artists.Items))
	for i, artist := range res.Artists.Items {
		result.Artists[i] = source.Artist(artist)
	}

	return result
}

func (source SpotifySource) Song(a spotify.TrackObject) Song {
	var s Song
	s.Provider = Spotify
	s.Album = source.Album(a.Album)
	s.Artists = make([]Artist, len(a.Artists))
	for i, a := range a.Artists {
		s.Artists[i] = source.Artist(a)
	}
	s.Name = a.Name
	s.URL = a.ExternalURLs.Spotify
	s.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Album.Images))
	s.Duration = time.Duration(a.DurationMS) * time.Millisecond
	s.ID = a.ID

	return s
}

func (source SpotifySource) Artist(a spotify.ArtistObject) Artist {
	var artist Artist
	artist.Name = a.Name
	artist.ID = a.ID
	artist.URL = a.ExternalURLs.Spotify
	artist.Listeners = a.Followers.Total
	artist.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Images))

	return artist
}

func (source SpotifySource) Album(a spotify.SimplifiedAlbumObject) Album {
	var album Album
	album.Name = a.Name
	album.ID = a.ID
	album.URL = a.ExternalURLs.Spotify
	album.Year, _ = strconv.Atoi(strings.Split(a.ReleaseDate, "-")[0])
	album.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Images))

	return album
}
