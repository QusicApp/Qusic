package lyrics

import (
	"github.com/qusicapp/qusic/spotify"
	"time"
)

func GetLyricsSpotify(client *spotify.Client, trackId string) (Song, error) {
	l, err := client.Lyrics(trackId)
	if err != nil {
		return Song{}, err
	}
	s := Song{
		LyricSource: "Spotify (" + l.ProviderDisplayName + ")",
	}
	switch l.SyncType {
	case "LINE_SYNCED":
		{
			s.SyncedLyrics = make([]SyncedLyric, len(l.Lines))
			for i, line := range l.Lines {
				s.SyncedLyrics[i] = SyncedLyric{
					At:    time.Duration(line.StartTimeMS.Int()) * time.Millisecond,
					Lyric: line.Words,
					Index: i,
				}
			}
		}
	case "UNSYNCED":
		{
			for i, line := range l.Lines {
				s.PlainLyrics += line.Words
				if i != len(l.Lines)-1 {
					s.PlainLyrics += "\n"
				}
			}
		}
	}

	return s, nil
}
