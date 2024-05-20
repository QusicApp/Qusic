package player

import (
	"fmt"
	"qusic/youtube"
	"unsafe"

	"fyne.io/fyne/v2"
)

type YouTubeMusicSource struct {
	client youtube.MusicClient
}

func (source YouTubeMusicSource) GetVideo(s *Song) {
	s.Video, _ = ytclient.GetVideo(s.URL)
	s.Format = &s.Video.Formats.Type("audio/webm; codecs=\"opus\"")[0]
}

func (source YouTubeMusicSource) Search(query string) SearchResult {
	var result SearchResult
	res, _ := source.client.Search(query)
	result.TopResult = source.Song(res.TopResult)
	showVideos := fyne.CurrentApp().Preferences().Bool("ytmusic.show_video_results")
	l := len(res.Songs)
	if showVideos {
		l += len(res.Videos)
	}

	result.Songs = make([]Song, l)
	for i, song := range res.Songs {
		result.Songs[i] = source.Song(song)
	}
	if showVideos {
		for in, vid := range res.Videos {
			i := in + len(res.Songs)
			result.Songs[i] = source.Song(vid)
		}
	}

	return result
}

func (source YouTubeMusicSource) Song(a youtube.Video) Song {
	var s Song
	s.Provider = YTMusic
	s.Album.Named = source.Named(a.Album, "https://music.youtube.com/browse")
	s.Artists = make([]Artist, len(a.Authors))
	for i, a := range a.Authors {
		s.Artists[i].Named = source.Named(a, "https://music.youtube.com/channel")
	}
	s.Name = a.Title
	s.URL = fmt.Sprintf("https://music.youtube.com/watch?v=%s", a.VideoID)
	s.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Thumbnails))
	s.Duration = a.Duration
	s.Plays = a.Plays
	s.ID = a.VideoID

	return s
}

func (source YouTubeMusicSource) Artist(a youtube.Artist) Artist {
	var artist Artist
	artist.Name = a.Name
	artist.ID = a.ID
	artist.URL = fmt.Sprintf("https://music.youtube.com/channel/%s", a.ID)
	artist.Listeners = a.Subscribers
	artist.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Thumbnails))

	return artist
}

func (source YouTubeMusicSource) Named(a youtube.Named, baseUrl string) Named {
	var album Named
	album.Name = a.Name
	album.ID = a.ID
	album.URL = fmt.Sprintf("%s/%s", baseUrl, a.ID)

	return album
}

func (source YouTubeMusicSource) Album(a youtube.Album) Album {
	var album Album
	album.Name = a.Name
	album.ID = a.ID
	album.URL = fmt.Sprintf("https://music.youtube.com/browse/%s", a.ID)
	album.Year = a.Year
	album.Thumbnails = *(*Thumbnails)(unsafe.Pointer(&a.Thumbnails))

	return album
}
