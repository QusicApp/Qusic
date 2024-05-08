package player

import (
	"qusic/spotify"
	"qusic/youtube"
	"strconv"
	"strings"
	"time"
	"unsafe"

	yt "github.com/kkdai/youtube/v2"
)

var ytclient = new(yt.Client)

func NewSpotifySource() SpotifySource {
	return SpotifySource{client: spotify.New()}
}

type SpotifySource struct {
	client *spotify.Client
}

func (source SpotifySource) GetVideo(s *Song) {
	v, _ := (*youtube.MusicClient).SearchSongs(nil, s.Artists[0].Name+" - "+s.Name)

	var vid *youtube.Video
	for _, video := range v {
		if abs(video.Duration-s.Duration) <= 2*time.Second {
			vid = &video
			break
		}
	}

	if vid == nil {
		return
	}

	video, err := ytclient.GetVideo(vid.VideoID)
	if err != nil {
		return
	}
	s.Video = video
	s.StreamURL = video.Formats.Type("audio")[0].URL
}

func (source SpotifySource) Search(query string) SearchResult {
	var result SearchResult
	res, _ := source.client.Search(query, spotify.QueryAll, "", nil, nil, false)
	if len(res.Tracks.Items) == 0 {
		return result
	}
	result.TopResult = source.Song(res.Tracks.Items[0])
	result.Songs = make([]Song, len(res.Tracks.Items[1:]))
	for i, song := range res.Tracks.Items[1:] {
		result.Songs[i] = source.Song(song)
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
