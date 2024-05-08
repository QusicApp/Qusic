package lyrics

import "qusic/youtube"

func GetSongYTMusic(videoId string) (Song, error) {
	lyrics, err := (*youtube.MusicClient).Lyrics(nil, videoId)
	return Song{
		LyricSource: "YouTube Music (" + lyrics.Source + ")",
		PlainLyrics: lyrics.Lyrics,
	}, err
}
