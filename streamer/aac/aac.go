package aac

// library to decode aac frames based on faad2

import (
	"bytes"

	"github.com/Eyevinn/mp4ff/bits"
)

func decode_fil(rd Reader) {
	count, _ := rd.Read(4)
	if count == 15 {
		c, _ := rd.Read(8)
		count += c
	}
	extension_type, _ := rd.Read(4)
	switch extension_type {
	case FILL_DATA:
		d, _ := rd.Read(4)
		if d == 0 {
			rd.Read((int(count) - 2) * 8)
		}
	default:
	}
}

func DecodeAACFrame(data []byte, sampleRateIndex, frameSizeFlag uint) /*(coef1, coef2 [1024]float64)*/ (samples0, samples1 []float64) {
	rd := Reader{
		bits.NewReader(bytes.NewReader(data)),
		sampleRateIndex,
		1024,
	}
	if frameSizeFlag == 1 {
		rd.frame_length = 960
	}
	elemType, _ := rd.Read(3)
	switch elemType {
	case CPE:
		coef1, coef2 := decode_cpe(rd)
		smp1, smp2 := make([]float64, 2048), make([]float64, 2048)

		imdct(new_mdct(rd.frame_length), coef1[:], smp1)
		imdct(new_mdct(rd.frame_length), coef2[:], smp2)

		return smp1, smp2
	default:
		return
	}
}
