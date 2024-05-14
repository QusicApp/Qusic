package streamer

import (
	"bytes"
	"io"
	"time"

	"github.com/ebml-go/webm"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/kkdai/youtube/v2"
	"layeh.com/gopus"
)

var yt youtube.Client

/*var _ beep.StreamSeekCloser = (*videoStreamer)(nil)

type videoStreamer struct {
	beep.StreamSeekCloser

	video *youtube.Video
	format *youtube.Format

	cmd *exec.Cmd
}

func (v videoStreamer) Seek(p int) (err error) {
	d := SampleRate.D(p)

	fmt.Println("seek to", d)
	v.cmd.Process.Kill()
	st, _, err := New(v.video, v.format, d)
	v = videoStreamer{
		StreamSeekCloser: st,

		video:  v.video,
		format: v.format,
	}
	return
}*/

var SampleRate beep.SampleRate = 48000

func New(video *youtube.Video, format *youtube.Format, at time.Duration) (beep.Streamer, beep.Format, error) {
	//path := fyne.CurrentApp().Preferences().StringWithFallback("ffmpeg_path", "ffmpeg")

	stream, _, err := yt.GetStream(video, format)

	if err != nil {
		return nil, beep.Format{}, err
	}

	data := make([]byte, format.ContentLength)
	if _, err := stream.Read(data); err != nil {
		return nil, beep.Format{}, err
	}

	/*dur :=
		fmt.Sprintf("%02d:%02d:%02d", int(at.Hours())%60, int(at.Minutes())%60, int(at.Seconds())%60)

	cmd := exec.Command(path,
		"-ss", dur,
		"-i",
		"pipe:0",
		"-f", "flac",
		"-ar", "44100",
		"-c:a", "flac",
		"-q:a", "4",
		"pipe:1",
	)
	cmd.Stdin = stream

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, beep.Format{}, err
	}

	if err := cmd.Start(); err != nil {
		return nil, beep.Format{}, err
	}

	streamer, fmt, err := flac.Decode(ReadCloser{bufio.NewReaderSize(pipe, 65307)})

	return videoStreamer{
		StreamSeekCloser: streamer,
		video:            video,
		format:           format,

		cmd: cmd,
	}, fmt, err*/
	var w webm.WebM
	reader, err := webm.Parse(bytes.NewReader(data), &w)
	if err != nil {
		return nil, beep.Format{}, err
	}

	audioTrack := w.FindFirstAudioTrack()
	form := beep.Format{
		SampleRate:  beep.SampleRate(audioTrack.SamplingFrequency),
		NumChannels: int(audioTrack.Channels),
		Precision:   2,
	}

	decoder, err := gopus.NewDecoder(int(audioTrack.SamplingFrequency), int(audioTrack.Channels))
	if err != nil {
		return nil, beep.Format{}, err
	}

	var frameSizeMs = 30.0
	var frameSize = float64(audioTrack.Channels) * frameSizeMs * audioTrack.SamplingFrequency / 1000

	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		frame := <-reader.Chan

		pcm, err := decoder.Decode(frame.Data, int(frameSize), false)
		if err != nil {
			return 0, false
		}
		for i := range samples {
			left := float64(pcm[i*2]) / (1 << 15)
			right := float64(pcm[i*2+1]) / (1 << 15)
			samples[i][0] = left
			samples[i][1] = right

			//fmt.Println(left, right)
		}
		return len(samples), true
	}), form, err
}

type ReadCloser struct {
	io.Reader
}

func (ReadCloser) Close() error {
	return nil
}

var _ beep.StreamSeekCloser = (*Streamer)(nil)

type Streamer struct {
	v effects.Volume
}

func NewStreamer() *Streamer {
	return &Streamer{
		v: effects.Volume{
			Base:     2,
			Streamer: new(beep.Ctrl),
		},
	}
}

func (s *Streamer) SetStreamer(st beep.Streamer) {
	s.v.Streamer.(*beep.Ctrl).Streamer = st
}

func (s *Streamer) SetPaused(v bool) {
	s.v.Streamer.(*beep.Ctrl).Paused = v
}

func (s *Streamer) Paused() bool {
	return s.v.Streamer.(*beep.Ctrl).Paused
}

func (s *Streamer) SetVolume(x float64) {
	s.v.Volume = x
}

func (s *Streamer) Volume() float64 {
	return s.v.Volume
}

func (s *Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	return s.v.Stream(samples)
}

func (s *Streamer) Err() error {
	return s.v.Err()
}

func (s *Streamer) Len() int {
	return 0 //s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Len()
}

func (s *Streamer) Position() int {
	return 0 //s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Position()
}

func (s *Streamer) Seek(p int) error {
	return nil //s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Seek(p)
}

func (s *Streamer) Close() error {
	return nil //s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Close()
}
