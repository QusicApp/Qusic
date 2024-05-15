package streamer

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/ebml-go/webm"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/kkdai/youtube/v2"
	"layeh.com/gopus"
)

var yt youtube.Client

var SampleRate beep.SampleRate = 48000

func New(video *youtube.Video) (beep.StreamSeekCloser, beep.Format, error) {
	format := video.Formats.Type("audio/webm; codecs=\"opus\"")[0]
	stream, _, err := yt.GetStream(video, &format)

	if err != nil {
		return nil, beep.Format{}, err
	}

	data := make([]byte, format.ContentLength)
	if _, err := stream.Read(data); err != nil {
		return nil, beep.Format{}, err
	}

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

	var frameSize = float64(audioTrack.Channels) * 60.0 * audioTrack.SamplingFrequency / 1000

	return &pcmStreamer{
		frameSize: int(frameSize),
		decoder:   decoder,
		packets:   reader.Chan,
		reader:    reader,

		duration: video.Duration,
	}, form, err
}

var ErrAlreadyClosed = errors.New("already closed")

type pcmStreamer struct {
	pcm       []int16
	pcmIdx    int
	frameSize int

	decoder *gopus.Decoder
	packets <-chan webm.Packet
	reader  *webm.Reader

	duration time.Duration

	pos int

	err error

	closed bool
}

var _ beep.StreamSeekCloser = (*pcmStreamer)(nil)

func (s *pcmStreamer) Err() error {
	err := s.err
	s.err = nil
	return err
}

func (s *pcmStreamer) Seek(n int) error {
	s.reader.Seek(SampleRate.D(n))
	s.pos = n
	s.pcm = nil
	s.pcmIdx = 0
	return nil
}

func (s *pcmStreamer) Position() int {
	return s.pos
}

func (s *pcmStreamer) Len() int {
	return SampleRate.N(s.duration)
}

func (s *pcmStreamer) Close() error {
	if s.closed {
		return ErrAlreadyClosed
	}
	s.closed = true
	return nil
}

func (s *pcmStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.closed {
		s.err = ErrAlreadyClosed
		return
	}
	if s.pos >= SampleRate.N(s.duration) {
		s.err = io.EOF
		return 0, false
	}
	for n < len(samples) {
		if s.pcmIdx >= len(s.pcm) {
			select {
			case packet := <-s.packets:
				s.pcm, s.err = s.decoder.Decode(packet.Data, int(s.frameSize), false)
				s.pcmIdx = 0
			default:
			}
		}
		if len(s.pcm) <= s.pcmIdx || s.err != nil {
			//audio done
			break
		}
		for ; n < len(samples); n++ {
			left := float64(s.pcm[s.pcmIdx]) / 32767
			right := float64(s.pcm[s.pcmIdx+1]) / 32767
			samples[n][0] = left
			samples[n][1] = right
			s.pcmIdx += 2
			if s.pcmIdx >= len(s.pcm) {
				break
			}
		}
	}

	s.pos += len(samples)

	return n, true
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

func (s *Streamer) SetStreamer(st beep.StreamSeekCloser) {
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
	if s.v.Streamer.(*beep.Ctrl).Streamer == nil {
		return s.v.Err()
	}
	return s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Err()
}

func (s *Streamer) Len() int {
	return s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Len()
}

func (s *Streamer) Position() int {
	return s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Position()
}

func (s *Streamer) Seek(p int) error {
	return s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Seek(p)
}

func (s *Streamer) Close() error {
	return s.v.Streamer.(*beep.Ctrl).Streamer.(beep.StreamSeekCloser).Close()
}
