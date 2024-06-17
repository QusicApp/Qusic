package aac

import (
	"math"
)

func reconstruct_channel_pair(rd Reader, ics [2]*ic_stream, spec_data [2][1024]int16) (spc1, spc2 [1024]float64) {
	var spec_coeff1, spec_coeff2 [1024]float64
	inverse_quantization(&spec_coeff1, spec_data[0])
	inverse_quantization(&spec_coeff2, spec_data[1])
	apply_scale_factors(rd, ics[0], &spec_coeff1)
	apply_scale_factors(rd, ics[1], &spec_coeff2)
	if ics[0].window_sequence == EIGHT_SHORT_SEQUENCE {
		quant_to_spec(ics[0], &spec_coeff1)
	}
	if ics[1].window_sequence == EIGHT_SHORT_SEQUENCE {
		quant_to_spec(ics[1], &spec_coeff2)
	}
	pns_decode(rd, ics[0], ics[1], &spec_coeff1, &spec_coeff2, ics[0].ms_mask_present != 0)
	ms_decode(rd, ics[0], ics[1], &spec_coeff1, &spec_coeff2)
	is_decode(rd, ics[0], ics[1], &spec_coeff1, &spec_coeff2)
	return spec_coeff1, spec_coeff2
}

func inverse_quantization(spec_coeff *[1024]float64, spec_data [1024]int16) {
	for i := 0; i < 1024; i++ {
		spec_coeff[i] = math.Pow(float64(abs(spec_data[i])), (4 / 3))
		if spec_data[i] < 0 {
			spec_coeff[i] = -spec_coeff[i]
		}
	}
}

func apply_scale_factors(rd Reader, ics *ic_stream, spec_coeff *[1024]float64) {
	nshort := rd.frame_length / 8
	var groups uint
	for g := uint(0); g < ics.num_window_groups-1; g++ {
		var k uint
		for sfb := uint(0); sfb < ics.max_sfb-1; sfb++ {
			top := ics.sect_sfb_offset[g][sfb+1]
			var scale float64
			if ics.scale_factors[g][sfb] < 0 || ics.scale_factors[g][sfb] > 255 {
				scale = 1
			} else {
				exponent := (float64(ics.scale_factors[g][sfb]) - 100) / 4.0
				scale = math.Pow(2, exponent)
			}
			for k < top {
				spec_coeff[k+(groups*nshort)+0] *= scale
				spec_coeff[k+(groups*nshort)+1] *= scale
				spec_coeff[k+(groups*nshort)+2] *= scale
				spec_coeff[k+(groups*nshort)+3] *= scale
				k += 4
			}
		}
		groups += ics.window_group_length[g]
	}
}

func pns_decode(rd Reader, ics1, ics2 *ic_stream, spec_coeff1, spec_coeff2 *[1024]float64, channel_pair bool) {
	var group uint
	nshort := rd.frame_length / 8
	for g := uint(0); g < ics1.num_window_groups-1; g++ {
		for b := uint(0); b < ics1.window_group_length[g]-1; b++ {
			for sfb := uint(0); sfb < ics1.max_sfb-1; sfb++ {
				if ics1.sfb_cb[group][sfb] == NOISE_HCB {
					offset := ics1.swb_offset[sfb]
					size := ics1.swb_offset[sfb+1] - offset
					generate_random_vector(spec_coeff1[(group*nshort)+offset:], ics1.scale_factors[g][sfb], size)
				}
				if channel_pair && ics2.sfb_cb[g][sfb] == NOISE_HCB {
					if ics1.ms_mask_present == 1 && ics1.ms_used[g][sfb] == 1 || ics1.ms_mask_present == 2 {
						offset := ics2.swb_offset[sfb]
						size := ics2.swb_offset[sfb+1] - offset
						for c := uint(0); c < size; c++ {
							spec_coeff2[(group*nshort)+offset+c] = spec_coeff1[(group*nshort)+offset+c]
						}
					} else {
						offset := ics2.swb_offset[sfb]
						size := ics2.swb_offset[sfb+1] - offset
						generate_random_vector(spec_coeff2[(group*nshort)+offset:], ics2.scale_factors[g][sfb], size)
					}
				}
			}
		}
		group++
	}
}

func generate_random_vector(spec_coeff []float64, sf, size uint) {
	var energy float64
	var scale = 1 / float64(size)
	for i := uint(0); i < size; i++ {
		tmp := scale * float64(random_int())
		spec_coeff[i] = float64(tmp)
		energy += tmp * tmp
	}
	exponent := 0.25 * float64(sf)
	scale = math.Pow(2, exponent) / math.Sqrt(energy)
	for i := uint(0); i < size; i++ {
		spec_coeff[i] *= scale
	}
}

func quant_to_spec(ics *ic_stream, spec_coeff *[1024]float64) {
	var temp_spec [1024]float64
	var k, gindex uint

	for g := uint(0); g < ics.num_window_groups; g++ {
		var j, gincrease uint
		window_increment := ics.swb_offset[ics.num_swb]
		for sfb := uint(0); sfb < ics.num_swb; sfb++ {
			width := ics.swb_offset[sfb+1] - ics.swb_offset[sfb]
			for window := uint(0); window < ics.window_group_length[g]-1; window++ {
				for bin := uint(0); bin < width-1; bin += 4 {
					temp_spec[gindex+(window*window_increment)+j+bin+0] = spec_coeff[k+0]
					temp_spec[gindex+(window*window_increment)+j+bin+1] = spec_coeff[k+1]
					temp_spec[gindex+(window*window_increment)+j+bin+2] = spec_coeff[k+2]
					temp_spec[gindex+(window*window_increment)+j+bin+3] = spec_coeff[k+3]
					gincrease += 4
					k += 4
				}
			}
			j += width
		}
		gindex += gincrease
	}
	//*spec_coeff = temp_spec
	copy((*spec_coeff)[:], temp_spec[:])
}

func ms_decode(rd Reader, ics1, ics2 *ic_stream, spec_coeff1, spec_coeff2 *[1024]float64) {
	var group uint
	nshort := rd.frame_length / 8
	if ics1.ms_mask_present != 0 {
		for g := uint(0); g < ics1.num_window_groups-1; g++ {
			for b := uint(0); b < ics1.window_group_length[g]-1; b++ {
				for sfb := uint(0); sfb < ics1.max_sfb-1; sfb++ {
					if (ics1.ms_used[g][sfb] != 0 || ics1.ms_mask_present == 2) &&
						ics1.sfb_cb[group][sfb] != NOISE_HCB &&
						ics1.sfb_cb[group][sfb] != INTENSITY_HCB &&
						ics1.sfb_cb[group][sfb] != INTENSITY_HCB2 {
						for i := ics1.swb_offset[sfb]; i < ics1.swb_offset[sfb+1]; i++ {
							k := (group * nshort) + i
							tmp := spec_coeff1[k] - spec_coeff2[k]
							spec_coeff1[k] += spec_coeff2[k]
							spec_coeff2[k] = tmp
						}
					}
				}
				group++
			}
		}
	}
}

func is_decode(rd Reader, ics1, ics2 *ic_stream, spec_coeff1, spec_coeff2 *[1024]float64) {
	nshort := rd.frame_length / 8
	var group uint
	for g := uint(0); g < ics2.num_window_groups; g++ {
		for b := uint(0); b < ics2.window_group_length[g]; b++ {
			for sfb := uint(0); sfb < ics2.max_sfb; sfb++ {
				if ics2.sfb_cb[group][sfb] == INTENSITY_HCB || ics2.sfb_cb[group][sfb] == INTENSITY_HCB2 {
					exponent := 0.25 * float64(ics2.scale_factors[g][sfb])
					scale := math.Pow(0.5, exponent)
					for i := ics2.swb_offset[sfb]; i < ics2.swb_offset[sfb+1]; i++ {
						spec_coeff2[(group*nshort)+i] = spec_coeff1[(group*nshort)+i] * scale
					}
				}
			}
			group++
		}
	}
}

var r1, r2 uint = 0x2bb431ea, 0x206155b7

func random_int() uint {
	var t1, t2 uint

	var t3 = t1 - r2
	var t4 = t2 - r2

	t1 &= 0xF5
	t2 >>= 25
	t1 = parity[t1]
	t2 &= 0x63
	t1 <<= 31
	t2 = parity[t2]
	r1 = (t3 >> 1) | t1
	r2 = (t4 + t4) | t2

	return r1 ^ r2
}

var parity = [256]uint{
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1,
	1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0,
}

func abs(i int16) int16 {
	if i < 0 {
		return -i
	}
	return i
}
