package cobalt

import (
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/jfbus/httprs"
	"github.com/kkdai/youtube/v2"
	"github.com/lostdusty/gobalt"
)

// TODO fix wav&mp3, fix seeking
func New(video *youtube.Video) (beep.StreamSeekCloser, beep.Format, error) {
	format := fyne.CurrentApp().Preferences().StringWithFallback("download.cobalt.format", "wav")
	settings := gobalt.Settings{
		Url:             "https://youtube.com/watch?v=" + video.ID,
		FilenamePattern: gobalt.Classic,
		AudioOnly:       true,
	}
	switch format {
	case "wav":
		settings.AudioCodec = gobalt.Wav
	case "mp3":
		settings.AudioCodec = gobalt.MP3
	case "ogg":
		settings.AudioCodec = gobalt.Ogg
	}
	file, err := gobalt.Run(settings)
	if err != nil {
		return nil, beep.Format{}, err
	}
	res, err := http.Get(file.URL)
	if err != nil {
		return nil, beep.Format{}, err
	}
	rd := httprs.NewHttpReadSeeker(res)
	switch format {
	case "wav":
		return wav.Decode(rd)
	case "mp3":
		return mp3.Decode(rd)
	case "ogg":
		return vorbis.Decode(rd)
	}
	return nil, beep.Format{}, nil
}

type readCloser struct {
	io.Reader
}

func (readCloser) Close() error {
	return nil
}
