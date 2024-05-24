package streamer

import (
	"bufio"
	"io"
	"os/exec"
	"qusic/spotify"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
)

// TODO add seeking (important)
func NewSpotifyMP4FlacStreamer(trackId string, client *spotify.Client) (beep.StreamSeekCloser, beep.Format, error) {
	buf, err := client.GetMP4(trackId)
	if err != nil {
		return nil, beep.Format{}, err
	}

	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-f", "flac",
		"-ar", "48000",
		"-q:a", "4",
		"pipe:1",
	)
	cmd.Stdin = buf

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, beep.Format{}, err
	}

	if err := cmd.Start(); err != nil {
		return nil, beep.Format{}, err
	}

	streamer, format, err := flac.Decode(ReadCloser{bufio.NewReaderSize(pipe, 65307)})
	if err != nil {
		return nil, beep.Format{}, err
	}

	return streamer, format, nil
}

type ReadCloser struct {
	io.Reader
}

func (ReadCloser) Close() error {
	return nil
}
