package aac

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

func DecodeAACFrame(data []byte, sampleRateIndex, frameSizeFlag uint) (coef1, coef2 [1024]float64) {
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
		return decode_cpe(rd)
	default:
		return
	}
}
