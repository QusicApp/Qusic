package streamer

import (
	"bytes"
	"fmt"
	"io"
	"qusic/preferences"
	"qusic/spotify"
	"qusic/streamer/aac"

	"github.com/Eyevinn/mp4ff/bits"
	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/gopxl/beep"
)

func NewSpotifyMP4AACStreamer(trackId string, client *spotify.Client) (beep.StreamSeekCloser, beep.Format, error) {
	buf, err := client.GetMP4(trackId)
	if err != nil {
		return nil, beep.Format{}, err
	}
	file, err := mp4.DecodeFile(buf)
	if err != nil {
		return nil, beep.Format{}, err
	}

	mp4a := file.Moov.Trak.Mdia.Minf.Stbl.Stsd.Mp4a
	format := beep.Format{
		SampleRate:  beep.SampleRate(mp4a.SampleRate),
		NumChannels: int(mp4a.ChannelCount),
	}
	audioConfig := mp4a.Esds.DecConfigDescriptor.DecSpecificInfo.DecConfig
	if len(audioConfig) != 2 {
		return nil, beep.Format{}, fmt.Errorf("audio config missing")
	}
	r := bits.NewReader(bytes.NewReader(audioConfig))
	trex := file.Moov.Mvex.Trex

	var samples []mp4.FullSample

	for i, box := range file.Children {
		if box.Type() == "moof" {
			moof := box.(*mp4.MoofBox)
			trun := moof.Traf.Trun

			mdat := file.Children[i+1].(*mp4.MdatBox)

			trun.AddSampleDefaultValues(moof.Traf.Tfhd, trex)

			samples = append(samples, trun.GetFullSamples(0, 0, mdat)...)
		}
	}

	if !preferences.Preferences.Bool("debug_mode") {
		return nil, beep.Format{}, fmt.Errorf("The Spotify mp4 decoder is not yet ready. In more technical info: I am making an AAC LC (Advanced Audio Coding Low Complexity) decoder, which is the audio format found in mp4s downloaded from Spotify. In the mean time, you can use the YouTube Music source or enable the \"Download songs from YouTube\" option in the settings, or wait for the decoder to be implemented.")
	}

	return &fmp4Streamer{
		frames:               samples,
		objectType:           r.MustRead(5),
		frequencyIndex:       r.MustRead(4),
		channelConfiguration: r.MustRead(4),
		frameLengthFlag:      r.MustRead(1),
		dependsOnCoreCoder:   r.MustRead(1),
		extensionFlag:        r.MustRead(1),
		decConfig:            mp4a.Esds.DecConfigDescriptor.DecSpecificInfo.DecConfig,
	}, format, nil
}

type fmp4Streamer struct {
	objectType, frequencyIndex, channelConfiguration, frameLengthFlag, dependsOnCoreCoder, extensionFlag uint

	decConfig []byte

	frames []mp4.FullSample

	samplesPerFrame int

	i int

	err error
}

func (f *fmp4Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	frame := f.frames[f.i]

	aac.DecodeAACFrame(frame.Data, f.frequencyIndex, f.frameLengthFlag)

	f.i++
	f.samplesPerFrame = len(samples)
	return len(samples), true
}
func (f fmp4Streamer) Err() error { return f.err }
func (fmp4Streamer) Close() error { return nil }
func (fmp4Streamer) Len() int     { return 0 }
func (f fmp4Streamer) Position() int {
	return f.i * f.samplesPerFrame
}
func (f fmp4Streamer) Seek(i int) error {
	f.i = i / f.samplesPerFrame
	return nil
}

type ReadCloser struct {
	io.Reader
}

func (ReadCloser) Close() error {
	return nil
}
