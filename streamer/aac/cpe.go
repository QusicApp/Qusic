package aac

import (
	"bytes"
	"fmt"

	"github.com/Eyevinn/mp4ff/bits"
)

type Reader struct {
	*bits.Reader
	sf_index, frame_length uint
}

func (r Reader) Skip(bits int) error {
	_, err := r.Read(bits)
	return err
}

type element struct {
	instance_tag  uint
	common_window bool
}

type ic struct {
	ics_index int

	element               element
	window_sequence       uint
	window_shape          uint
	scale_factor_grouping uint
	max_sfb               uint

	num_sec             [8]uint
	num_swb             uint
	num_windows         uint
	num_window_groups   uint
	window_group_length [8]uint

	ms_used, scale_factors [8][51]uint

	swb_offset [52]uint

	sect_sfb_offset,
	sect_cb, sect_start, sect_end, sfb_cb [8][15 * 8]uint

	predictor_data_present bool

	global_gain uint

	noise_used bool

	ms_mask_present uint
}

func decode_cpe(rd Reader) [2][1024]int16 {
	var element element
	element.instance_tag, _ = rd.Read(4)
	element.common_window, _ = rd.ReadFlag()
	var (
		ics1, ics2             ic
		spec_data1, spec_data2 [1024]int16
	)
	ics1.element = element
	ics2.element = element

	if element.common_window {
		ics_info(rd, &ics1)
		ics1.ms_mask_present, _ = rd.Read(2)
		if ics1.ms_mask_present == 1 {
			for g := uint(0); g < ics1.num_window_groups; g++ {
				for sfb := uint(0); sfb < ics1.max_sfb; sfb++ {
					ics1.ms_used[g][sfb], _ = rd.Read(1)
				}
			}
		}
		ics2 = ics1
	} else {
		fmt.Println("not common window")
		ics1.ms_mask_present = 0
	}
	ics2.ics_index = 1
	individual_channel_stream(rd, &ics1, &spec_data1)
	individual_channel_stream(rd, &ics2, &spec_data2)

	return [2][1024]int16{spec_data1, spec_data2}
}

func ics_info(rd Reader, ics *ic) {
	rd.Read(1) // reserved
	ics.window_sequence, _ = rd.Read(2)
	ics.window_shape, _ = rd.Read(1)

	if ics.window_sequence == EIGHT_SHORT_SEQUENCE {
		ics.max_sfb, _ = rd.Read(4)
		ics.scale_factor_grouping, _ = rd.Read(7)
	} else {
		ics.max_sfb, _ = rd.Read(6)
	}
	window_grouping_info(rd, ics)

	if ics.window_sequence != EIGHT_SHORT_SEQUENCE {
		ics.predictor_data_present, _ = rd.ReadFlag()
	}
}

func window_grouping_info(rd Reader, ics *ic) {
	switch ics.window_sequence {
	case ONLY_LONG_SEQUENCE, LONG_START_SEQUENCE, LONG_STOP_SEQUENCE:
		ics.num_windows = 1
		ics.num_window_groups = 1
		ics.window_group_length[ics.num_window_groups-1] = 1

		if rd.frame_length == 1024 {
			ics.num_swb = num_swb_1024_window[rd.sf_index]
		} else {
			ics.num_swb = num_swb_960_window[rd.sf_index]
		}
		for i := uint(0); i < ics.num_swb; i++ {
			ics.sect_sfb_offset[0][i] = swb_offset_1024_window[rd.sf_index][i]
			ics.swb_offset[i] = swb_offset_1024_window[rd.sf_index][i]
		}
		ics.sect_sfb_offset[0][ics.num_swb] = rd.frame_length
		ics.swb_offset[ics.num_swb] = rd.frame_length
	case EIGHT_SHORT_SEQUENCE:
		fmt.Println("short")
	}
}

func individual_channel_stream(rd Reader, ics *ic, spec_data *[1024]int16) {
	ics.global_gain, _ = rd.Read(8)
	if !ics.element.common_window {
		ics_info(rd, ics)
	}
	section_data(rd, ics)
	decode_scale_factors(rd, ics)
	pulse_data_present, _ := rd.ReadFlag()
	if pulse_data_present {
		number_pulse, _ := rd.Read(2)
		rd.Read(6 + (int(number_pulse) * 9))
	}
	tns_data_present, _ := rd.ReadFlag()
	if tns_data_present {
		var (
			n_filter_bits = 2
			length_bits   = 4
			order_bits    = 3
		)
		if ics.window_sequence == EIGHT_SHORT_SEQUENCE {
			n_filter_bits = 1
			length_bits = 4
			order_bits = 3
		}
		for w := uint(0); w < ics.num_windows; w++ {
			var start_coef_bits uint = 3
			n_filter, _ := rd.Read(n_filter_bits)
			if n_filter != 0 {
				coef_res, _ := rd.ReadFlag()
				if coef_res {
					start_coef_bits = 4
				}
			}
			for filter := uint(0); filter < n_filter; filter++ {
				_, _ = rd.Read(length_bits) // length
				order, _ := rd.Read(order_bits)
				if order != 0 {
					_, _ = rd.Read(1) // direction
					coef_compress, _ := rd.Read(1)
					coefficient_bits := start_coef_bits - coef_compress
					for i := uint(0); i < order; i++ {
						rd.Read(int(coefficient_bits))
					}
				}
			}
		}
	}
	rd.ReadFlag() // gain_control_data_present
	spectral_data(rd, ics, spec_data)
}

func spectral_data(rd Reader, ics *ic, spec_data *[1024]int16) {
	var (
		p      uint
		groups uint = 0
		nshort      = rd.frame_length / 8
	)
	for g := uint(0); g < ics.num_window_groups; g++ {
		p = nshort * groups
		for i := uint(0); i < ics.num_sec[g]; i++ {
			section_codebook := ics.sect_cb[g][i]
			var increment uint = 4
			if section_codebook >= 5 {
				increment = 2
			}
			switch section_codebook {
			case ZERO_HCB, NOISE_HCB, INTENSITY_HCB, INTENSITY_HCB2:
				p += ics.sect_sfb_offset[g][ics.sect_end[g][i]] - ics.sect_sfb_offset[g][ics.sect_start[g][i]]
			default:
				for k := ics.sect_sfb_offset[g][ics.sect_start[g][i]]; k < ics.sect_sfb_offset[g][ics.sect_end[g][i]]; k++ {
					k += increment
				}
				huffman_spectral_data(section_codebook, rd, spec_data[p:])
				p += increment
			}
		}
		groups += ics.window_group_length[g]
	}
}

func huffman_scale_factor(rd Reader) uint {
	var (
		offset uint
	)
	for hcb_sf[offset][1] != 0 {
		b, _ := rd.Read(1)
		offset += hcb_sf[offset][b]
	}

	return hcb_sf[offset][0]
}

func decode_scale_factors(rd Reader, ics *ic) {
	var (
		g, sfb, t      uint
		noise_pcm_flag = true
		scale_factor   = ics.global_gain
		is_position    uint
		noise_energy   = ics.global_gain - 90
	)
	for g = 0; g < ics.num_window_groups; g++ {
		for sfb = 0; sfb < ics.max_sfb; sfb++ {
			switch ics.sfb_cb[g][sfb] {
			case ZERO_HCB:
				ics.scale_factors[g][sfb] = 0
			case INTENSITY_HCB, INTENSITY_HCB2:
				t = huffman_scale_factor(rd)
				is_position += (t - 60)
				ics.scale_factors[g][sfb] = is_position
			case NOISE_HCB:
				fmt.Printf("sfb:%d<noise>\n", sfb)
				if noise_pcm_flag {
					noise_pcm_flag = false
					t, _ = rd.Read(9)
				} else {
					t = huffman_scale_factor(rd)
					t -= 60
				}
				noise_energy += t
				ics.scale_factors[g][sfb] = noise_energy
			default:
				ics.scale_factors[g][sfb] = 0
				t = huffman_scale_factor(rd)
				scale_factor += (t - 60)
				if scale_factor < 0 || scale_factor > 255 {
					//error
					return
				}
				ics.scale_factors[g][sfb] = scale_factor
			}
		}
	}
}

func section_data(rd Reader, ics *ic) {
	var section_bits = 5
	if ics.window_sequence == EIGHT_SHORT_SEQUENCE {
		section_bits = 3
	}
	section_escape_value := uint(1<<section_bits) - 1

	for g := uint(0); g < ics.num_window_groups; g++ {
		var (
			k, i uint
		)
		for k < ics.max_sfb {
			var section_length uint
			ics.sect_cb[g][i], _ = rd.Read(4)
			if ics.sect_cb[g][i] == NOISE_HCB {
				ics.noise_used = true
			}
			section_length_increment, _ := rd.Read(section_bits)
			for section_length_increment == section_escape_value {
				section_length += section_length_increment
				section_length_increment, _ = rd.Read(section_bits)
			}
			section_length += section_length_increment
			ics.sect_start[g][i] = k
			ics.sect_end[g][i] = k + section_length
			if k+section_length >= 8*15 {
				fmt.Println("error0")
				//error
				return
			}
			if i >= 8*15 {
				fmt.Println("error1")
				//error
				return
			}
			for sfb := k; sfb < k+section_length; sfb++ {
				ics.sfb_cb[g][sfb] = ics.sect_cb[g][i]
			}
			k += section_length
			i++
		}
		ics.num_sec[g] = i
	}
}

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

func DecodeAACFrame(data []byte, sampleRateIndex, frameSizeFlag uint) [2][1024]int16 {
	rd := Reader{
		bits.NewReader(bytes.NewReader(data)),
		sampleRateIndex,
		1024,
	}
	if frameSizeFlag == 1 {
		rd.frame_length = 960
	}
	var spec [2][1024]int16
	for {
		elemType, err := rd.Read(3)
		if err != nil {
			break
		}
		switch elemType {
		case CPE:
			spec = decode_cpe(rd)
		case FIL:
			decode_fil(rd)
		case END:
			break
		}
	}
	return spec
}
