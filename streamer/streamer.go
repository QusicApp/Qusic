package streamer

import (
	"github.com/qusicapp/qusic/util"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/kkdai/youtube/v2"
)

const (
	ErrAlreadyClosed      util.StringError = "already closed"
	ErrClosed             util.StringError = "closed"
	ErrAudioConfigMissing util.StringError = "audio config missing"
)

var yt youtube.Client

var SampleRate beep.SampleRate = 48000

var _ beep.StreamSeekCloser = (*Streamer)(nil)

type Streamer struct {
	v *effects.Volume
	r *beep.Resampler
}

func NewStreamer() *Streamer {
	vol := &effects.Volume{
		Base:     2,
		Streamer: new(beep.Ctrl),
	}
	return &Streamer{
		v: vol,
		r: beep.ResampleRatio(4, 1, vol),
	}
}

func (s *Streamer) Playing() bool {
	return s.v.Streamer.(*beep.Ctrl).Streamer != nil
}

func (s *Streamer) SetStreamer(st beep.StreamSeekCloser) {
	s.v.Streamer.(*beep.Ctrl).Streamer = st
}

func (s *Streamer) SetRatio(ratio float64) {
	s.r.SetRatio(ratio)
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

func (s *Streamer) SetMute(b bool) {
	s.v.Silent = b
}

func (s *Streamer) Mute() bool {
	return s.v.Silent
}

func (s *Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	return s.r.Stream(samples)
}

func (s *Streamer) Err() error {
	if err := s.r.Err(); err != nil {
		return err
	}
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
