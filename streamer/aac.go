package streamer

import (
	"bytes"
	"fmt"

	"github.com/Eyevinn/mp4ff/bits"
)

const (
	SCE = iota
	CPE
	CCE
	LFE
	DSE
	PCE
	FIL
	END
)

const (
	ONLY_LONG_SEQUENCE = iota
	LONG_START_SEQUENCE
	EIGHT_SHORT_SEQUENCE
	LONG_STOP_SEQUENCE
)

const (
	ZERO_HCB       = 0
	FIRST_PAIR_HCB = 5
	ESC_HCB        = 11
	QUAD_LEN       = 4
	PAIR_LEN       = 2
	NOISE_HCB      = 13
	INTENSITY_HCB2 = 14
	INTENSITY_HCB  = 15
)

var (
	num_swb_1024_window = []uint{
		41, 41, 47, 49, 49, 51, 47, 47, 43, 43, 43, 40,
	}
	num_swb_960_window = []uint{
		40, 40, 45, 49, 49, 49, 46, 46, 42, 42, 42, 40,
	}

	swb_offset_1024_96 = []uint{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56,
		64, 72, 80, 88, 96, 108, 120, 132, 144, 156, 172, 188, 212, 240,
		276, 320, 384, 448, 512, 576, 640, 704, 768, 832, 896, 960, 1024,
	}
	swb_offset_1024_64 = []uint{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56,
		64, 72, 80, 88, 100, 112, 124, 140, 156, 172, 192, 216, 240, 268,
		304, 344, 384, 424, 464, 504, 544, 584, 624, 664, 704, 744, 784, 824,
		864, 904, 944, 984, 1024,
	}

	swb_offset_1024_48 = []uint{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 48, 56, 64, 72,
		80, 88, 96, 108, 120, 132, 144, 160, 176, 196, 216, 240, 264, 292,
		320, 352, 384, 416, 448, 480, 512, 544, 576, 608, 640, 672, 704, 736,
		768, 800, 832, 864, 896, 928, 1024,
	}

	swb_offset_1024_32 = []uint{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 48, 56, 64, 72,
		80, 88, 96, 108, 120, 132, 144, 160, 176, 196, 216, 240, 264, 292,
		320, 352, 384, 416, 448, 480, 512, 544, 576, 608, 640, 672, 704, 736,
		768, 800, 832, 864, 896, 928, 960, 992, 1024,
	}

	swb_offset_1024_24 = []uint{
		0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 52, 60, 68,
		76, 84, 92, 100, 108, 116, 124, 136, 148, 160, 172, 188, 204, 220,
		240, 260, 284, 308, 336, 364, 396, 432, 468, 508, 552, 600, 652, 704,
		768, 832, 896, 960, 1024,
	}

	swb_offset_1024_16 = []uint{
		0, 8, 16, 24, 32, 40, 48, 56, 64, 72, 80, 88, 100, 112, 124,
		136, 148, 160, 172, 184, 196, 212, 228, 244, 260, 280, 300, 320, 344,
		368, 396, 424, 456, 492, 532, 572, 616, 664, 716, 772, 832, 896, 960, 1024,
	}

	swb_offset_1024_8 = []uint{
		0, 12, 24, 36, 48, 60, 72, 84, 96, 108, 120, 132, 144, 156, 172,
		188, 204, 220, 236, 252, 268, 288, 308, 328, 348, 372, 396, 420, 448,
		476, 508, 544, 580, 620, 664, 712, 764, 820, 880, 944, 1024,
	}

	swb_offset_1024_window = [][]uint{
		swb_offset_1024_96, /* 96000 */
		swb_offset_1024_96, /* 88200 */
		swb_offset_1024_64, /* 64000 */
		swb_offset_1024_48, /* 48000 */
		swb_offset_1024_48, /* 44100 */
		swb_offset_1024_32, /* 32000 */
		swb_offset_1024_24, /* 24000 */
		swb_offset_1024_24, /* 22050 */
		swb_offset_1024_16, /* 16000 */
		swb_offset_1024_16, /* 12000 */
		swb_offset_1024_16, /* 11025 */
		swb_offset_1024_8,  /* 8000  */
	}
)

var aac_scalefactor_huffman_table = [][3]uint{
	/* codeword, code length, scalefactor */
	{0x0, 1, 60},
	{0x4, 3, 59},
	{0xA, 4, 61},
	{0xB, 4, 58},
	{0xC, 4, 62},
	{0x1A, 5, 57},
	{0x1B, 5, 63},
	{0x38, 6, 56},
	{0x39, 6, 64},
	{0x3A, 6, 55},
	{0x3B, 6, 65},
	{0x78, 7, 66},
	{0x79, 7, 54},
	{0x7A, 7, 67},
	{0xF6, 8, 53},
	{0xF7, 8, 68},
	{0xF8, 8, 52},
	{0xF9, 8, 69},
	{0xFA, 8, 51},
	{0x1F6, 9, 70},
	{0x1F7, 9, 50},
	{0x1F8, 9, 49},
	{0x1F9, 9, 71},
	{0x3F4, 10, 72},
	{0x3F5, 10, 48},
	{0x3F6, 10, 73},
	{0x3F7, 10, 47},
	{0x3F8, 10, 74},
	{0x3F9, 10, 46},
	{0x7F4, 11, 76},
	{0x7F5, 11, 75},
	{0x7F6, 11, 77},
	{0x7F7, 11, 78},
	{0x7F8, 11, 45},
	{0x7F9, 11, 43},
	{0xFF4, 12, 44},
	{0xFF5, 12, 79},
	{0xFF6, 12, 42},
	{0xFF7, 12, 41},
	{0xFF8, 12, 80},
	{0xFF9, 12, 40},
	{0x1FF4, 13, 81},
	{0x1FF5, 13, 39},
	{0x1FF6, 13, 82},
	{0x1FF7, 13, 38},
	{0x1FF8, 13, 83},
	{0x3FF2, 14, 37},
	{0x3FF3, 14, 35},
	{0x3FF4, 14, 85},
	{0x3FF5, 14, 33},
	{0x3FF6, 14, 36},
	{0x3FF7, 14, 34},
	{0x3FF8, 14, 84},
	{0x3FF9, 14, 32},
	{0x7FF4, 15, 87},
	{0x7FF5, 15, 89},
	{0x7FF6, 15, 30},
	{0x7FF7, 15, 31},
	{0xFFF0, 16, 86},
	{0xFFF1, 16, 29},
	{0xFFF2, 16, 26},
	{0xFFF3, 16, 27},
	{0xFFF4, 16, 28},
	{0xFFF5, 16, 24},
	{0xFFF6, 16, 88},
	{0x1FFEE, 17, 25},
	{0x1FFEF, 17, 22},
	{0x1FFF0, 17, 23},
	{0x3FFE2, 18, 90},
	{0x3FFE3, 18, 21},
	{0x3FFE4, 18, 19},
	{0x3FFE5, 18, 3},
	{0x3FFE6, 18, 1},
	{0x3FFE7, 18, 2},
	{0x3FFE8, 18, 0},
	{0x7FFD2, 19, 98},
	{0x7FFD3, 19, 99},
	{0x7FFD4, 19, 100},
	{0x7FFD5, 19, 101},
	{0x7FFD6, 19, 102},
	{0x7FFD7, 19, 117},
	{0x7FFD8, 19, 97},
	{0x7FFD9, 19, 91},
	{0x7FFDA, 19, 92},
	{0x7FFDB, 19, 93},
	{0x7FFDC, 19, 94},
	{0x7FFDD, 19, 95},
	{0x7FFDE, 19, 96},
	{0x7FFDF, 19, 104},
	{0x7FFE0, 19, 111},
	{0x7FFE1, 19, 112},
	{0x7FFE2, 19, 113},
	{0x7FFE3, 19, 114},
	{0x7FFE4, 19, 115},
	{0x7FFE5, 19, 116},
	{0x7FFE6, 19, 110},
	{0x7FFE7, 19, 105},
	{0x7FFE8, 19, 106},
	{0x7FFE9, 19, 107},
	{0x7FFEA, 19, 108},
	{0x7FFEB, 19, 109},
	{0x7FFEC, 19, 118},
	{0x7FFED, 19, 6},
	{0x7FFEE, 19, 8},
	{0x7FFEF, 19, 9},
	{0x7FFF0, 19, 10},
	{0x7FFF1, 19, 5},
	{0x7FFF2, 19, 103},
	{0x7FFF3, 19, 120},
	{0x7FFF4, 19, 119},
	{0x7FFF5, 19, 4},
	{0x7FFF6, 19, 7},
	{0x7FFF7, 19, 15},
	{0x7FFF8, 19, 16},
	{0x7FFF9, 19, 18},
	{0x7FFFA, 19, 20},
	{0x7FFFB, 19, 17},
	{0x7FFFC, 19, 11},
	{0x7FFFD, 19, 12},
	{0x7FFFE, 19, 14},
	{0x7FFFF, 19, 13},
}

var hcb_sf = [241][2]uint{
	{ /*   0 */ 1, 2},
	{ /*   1 */ 60, 0},
	{ /*   2 */ 1, 2},
	{ /*   3 */ 2, 3},
	{ /*   4 */ 3, 4},
	{ /*   5 */ 59, 0},
	{ /*   6 */ 3, 4},
	{ /*   7 */ 4, 5},
	{ /*   8 */ 5, 6},
	{ /*   9 */ 61, 0},
	{ /*  10 */ 58, 0},
	{ /*  11 */ 62, 0},
	{ /*  12 */ 3, 4},
	{ /*  13 */ 4, 5},
	{ /*  14 */ 5, 6},
	{ /*  15 */ 57, 0},
	{ /*  16 */ 63, 0},
	{ /*  17 */ 4, 5},
	{ /*  18 */ 5, 6},
	{ /*  19 */ 6, 7},
	{ /*  20 */ 7, 8},
	{ /*  21 */ 56, 0},
	{ /*  22 */ 64, 0},
	{ /*  23 */ 55, 0},
	{ /*  24 */ 65, 0},
	{ /*  25 */ 4, 5},
	{ /*  26 */ 5, 6},
	{ /*  27 */ 6, 7},
	{ /*  28 */ 7, 8},
	{ /*  29 */ 66, 0},
	{ /*  30 */ 54, 0},
	{ /*  31 */ 67, 0},
	{ /*  32 */ 5, 6},
	{ /*  33 */ 6, 7},
	{ /*  34 */ 7, 8},
	{ /*  35 */ 8, 9},
	{ /*  36 */ 9, 10},
	{ /*  37 */ 53, 0},
	{ /*  38 */ 68, 0},
	{ /*  39 */ 52, 0},
	{ /*  40 */ 69, 0},
	{ /*  41 */ 51, 0},
	{ /*  42 */ 5, 6},
	{ /*  43 */ 6, 7},
	{ /*  44 */ 7, 8},
	{ /*  45 */ 8, 9},
	{ /*  46 */ 9, 10},
	{ /*  47 */ 70, 0},
	{ /*  48 */ 50, 0},
	{ /*  49 */ 49, 0},
	{ /*  50 */ 71, 0},
	{ /*  51 */ 6, 7},
	{ /*  52 */ 7, 8},
	{ /*  53 */ 8, 9},
	{ /*  54 */ 9, 10},
	{ /*  55 */ 10, 11},
	{ /*  56 */ 11, 12},
	{ /*  57 */ 72, 0},
	{ /*  58 */ 48, 0},
	{ /*  59 */ 73, 0},
	{ /*  60 */ 47, 0},
	{ /*  61 */ 74, 0},
	{ /*  62 */ 46, 0},
	{ /*  63 */ 6, 7},
	{ /*  64 */ 7, 8},
	{ /*  65 */ 8, 9},
	{ /*  66 */ 9, 10},
	{ /*  67 */ 10, 11},
	{ /*  68 */ 11, 12},
	{ /*  69 */ 76, 0},
	{ /*  70 */ 75, 0},
	{ /*  71 */ 77, 0},
	{ /*  72 */ 78, 0},
	{ /*  73 */ 45, 0},
	{ /*  74 */ 43, 0},
	{ /*  75 */ 6, 7},
	{ /*  76 */ 7, 8},
	{ /*  77 */ 8, 9},
	{ /*  78 */ 9, 10},
	{ /*  79 */ 10, 11},
	{ /*  80 */ 11, 12},
	{ /*  81 */ 44, 0},
	{ /*  82 */ 79, 0},
	{ /*  83 */ 42, 0},
	{ /*  84 */ 41, 0},
	{ /*  85 */ 80, 0},
	{ /*  86 */ 40, 0},
	{ /*  87 */ 6, 7},
	{ /*  88 */ 7, 8},
	{ /*  89 */ 8, 9},
	{ /*  90 */ 9, 10},
	{ /*  91 */ 10, 11},
	{ /*  92 */ 11, 12},
	{ /*  93 */ 81, 0},
	{ /*  94 */ 39, 0},
	{ /*  95 */ 82, 0},
	{ /*  96 */ 38, 0},
	{ /*  97 */ 83, 0},
	{ /*  98 */ 7, 8},
	{ /*  99 */ 8, 9},
	{ /* 100 */ 9, 10},
	{ /* 101 */ 10, 11},
	{ /* 102 */ 11, 12},
	{ /* 103 */ 12, 13},
	{ /* 104 */ 13, 14},
	{ /* 105 */ 37, 0},
	{ /* 106 */ 35, 0},
	{ /* 107 */ 85, 0},
	{ /* 108 */ 33, 0},
	{ /* 109 */ 36, 0},
	{ /* 110 */ 34, 0},
	{ /* 111 */ 84, 0},
	{ /* 112 */ 32, 0},
	{ /* 113 */ 6, 7},
	{ /* 114 */ 7, 8},
	{ /* 115 */ 8, 9},
	{ /* 116 */ 9, 10},
	{ /* 117 */ 10, 11},
	{ /* 118 */ 11, 12},
	{ /* 119 */ 87, 0},
	{ /* 120 */ 89, 0},
	{ /* 121 */ 30, 0},
	{ /* 122 */ 31, 0},
	{ /* 123 */ 8, 9},
	{ /* 124 */ 9, 10},
	{ /* 125 */ 10, 11},
	{ /* 126 */ 11, 12},
	{ /* 127 */ 12, 13},
	{ /* 128 */ 13, 14},
	{ /* 129 */ 14, 15},
	{ /* 130 */ 15, 16},
	{ /* 131 */ 86, 0},
	{ /* 132 */ 29, 0},
	{ /* 133 */ 26, 0},
	{ /* 134 */ 27, 0},
	{ /* 135 */ 28, 0},
	{ /* 136 */ 24, 0},
	{ /* 137 */ 88, 0},
	{ /* 138 */ 9, 10},
	{ /* 139 */ 10, 11},
	{ /* 140 */ 11, 12},
	{ /* 141 */ 12, 13},
	{ /* 142 */ 13, 14},
	{ /* 143 */ 14, 15},
	{ /* 144 */ 15, 16},
	{ /* 145 */ 16, 17},
	{ /* 146 */ 17, 18},
	{ /* 147 */ 25, 0},
	{ /* 148 */ 22, 0},
	{ /* 149 */ 23, 0},
	{ /* 150 */ 15, 16},
	{ /* 151 */ 16, 17},
	{ /* 152 */ 17, 18},
	{ /* 153 */ 18, 19},
	{ /* 154 */ 19, 20},
	{ /* 155 */ 20, 21},
	{ /* 156 */ 21, 22},
	{ /* 157 */ 22, 23},
	{ /* 158 */ 23, 24},
	{ /* 159 */ 24, 25},
	{ /* 160 */ 25, 26},
	{ /* 161 */ 26, 27},
	{ /* 162 */ 27, 28},
	{ /* 163 */ 28, 29},
	{ /* 164 */ 29, 30},
	{ /* 165 */ 90, 0},
	{ /* 166 */ 21, 0},
	{ /* 167 */ 19, 0},
	{ /* 168 */ 3, 0},
	{ /* 169 */ 1, 0},
	{ /* 170 */ 2, 0},
	{ /* 171 */ 0, 0},
	{ /* 172 */ 23, 24},
	{ /* 173 */ 24, 25},
	{ /* 174 */ 25, 26},
	{ /* 175 */ 26, 27},
	{ /* 176 */ 27, 28},
	{ /* 177 */ 28, 29},
	{ /* 178 */ 29, 30},
	{ /* 179 */ 30, 31},
	{ /* 180 */ 31, 32},
	{ /* 181 */ 32, 33},
	{ /* 182 */ 33, 34},
	{ /* 183 */ 34, 35},
	{ /* 184 */ 35, 36},
	{ /* 185 */ 36, 37},
	{ /* 186 */ 37, 38},
	{ /* 187 */ 38, 39},
	{ /* 188 */ 39, 40},
	{ /* 189 */ 40, 41},
	{ /* 190 */ 41, 42},
	{ /* 191 */ 42, 43},
	{ /* 192 */ 43, 44},
	{ /* 193 */ 44, 45},
	{ /* 194 */ 45, 46},
	{ /* 195 */ 98, 0},
	{ /* 196 */ 99, 0},
	{ /* 197 */ 100, 0},
	{ /* 198 */ 101, 0},
	{ /* 199 */ 102, 0},
	{ /* 200 */ 117, 0},
	{ /* 201 */ 97, 0},
	{ /* 202 */ 91, 0},
	{ /* 203 */ 92, 0},
	{ /* 204 */ 93, 0},
	{ /* 205 */ 94, 0},
	{ /* 206 */ 95, 0},
	{ /* 207 */ 96, 0},
	{ /* 208 */ 104, 0},
	{ /* 209 */ 111, 0},
	{ /* 210 */ 112, 0},
	{ /* 211 */ 113, 0},
	{ /* 212 */ 114, 0},
	{ /* 213 */ 115, 0},
	{ /* 214 */ 116, 0},
	{ /* 215 */ 110, 0},
	{ /* 216 */ 105, 0},
	{ /* 217 */ 106, 0},
	{ /* 218 */ 107, 0},
	{ /* 219 */ 108, 0},
	{ /* 220 */ 109, 0},
	{ /* 221 */ 118, 0},
	{ /* 222 */ 6, 0},
	{ /* 223 */ 8, 0},
	{ /* 224 */ 9, 0},
	{ /* 225 */ 10, 0},
	{ /* 226 */ 5, 0},
	{ /* 227 */ 103, 0},
	{ /* 228 */ 120, 0},
	{ /* 229 */ 119, 0},
	{ /* 230 */ 4, 0},
	{ /* 231 */ 7, 0},
	{ /* 232 */ 15, 0},
	{ /* 233 */ 16, 0},
	{ /* 234 */ 18, 0},
	{ /* 235 */ 20, 0},
	{ /* 236 */ 17, 0},
	{ /* 237 */ 11, 0},
	{ /* 238 */ 12, 0},
	{ /* 239 */ 14, 0},
	{ /* 240 */ 13, 0},
}

type Reader struct {
	*bits.Reader
	sf_index, frame_length uint
}

type element struct {
	common_window bool
}

type ic struct {
	element               element
	window_sequence       uint
	window_shape          uint
	scale_factor_grouping uint
	max_sfb               uint

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
	window_sequence(rd, ics)

	if ics.window_sequence != EIGHT_SHORT_SEQUENCE {
		ics.predictor_data_present, _ = rd.ReadFlag()
	}
}

func window_sequence(rd Reader, ics *ic) {
	switch ics.window_sequence {
	case ONLY_LONG_SEQUENCE, LONG_START_SEQUENCE, LONG_STOP_SEQUENCE:
		fmt.Println("long")
		ics.num_windows = 1
		ics.num_window_groups = 1
		ics.window_group_length[ics.num_window_groups-1] = 1

		if rd.frame_length == 1024 {
			ics.num_swb = num_swb_1024_window[rd.sf_index]
		} else {
			ics.num_swb = num_swb_960_window[rd.sf_index]
		}
		for i := uint(0); i < ics.num_swb-1; i++ {
			ics.sect_sfb_offset[0][i] = swb_offset_1024_window[rd.sf_index][i]
			ics.swb_offset[i] = swb_offset_1024_window[rd.sf_index][i]
		}
		ics.sect_sfb_offset[0][ics.num_swb] = rd.frame_length
		ics.swb_offset[ics.num_swb] = rd.frame_length
	case EIGHT_SHORT_SEQUENCE:
		fmt.Println("short")
	}
}

func channel_pair_element(rd Reader) {
	var element element
	var (
		_, _ = rd.Read(4) // element_instance_tag
	)
	element.common_window, _ = rd.ReadFlag()
	var (
		ics1, ics2             ic
		spec_data1, spec_data2 [1024]int16
	)
	ics1.element = element
	ics2.element = element

	if element.common_window {
		fmt.Println("common window")
		ics_info(rd, &ics1)
		ics1.ms_mask_present, _ = rd.Read(2)
		if ics1.ms_mask_present == 1 {
			for g := uint(0); g < ics1.num_window_groups-1; g++ {
				for sfb := uint(0); sfb < ics1.max_sfb-1; sfb++ {
					ics1.ms_used[g][sfb], _ = rd.Read(1)
				}
			}
		}
		ics2 = ics1
	} else {
		fmt.Println("not common window")
		ics1.ms_mask_present = 0
	}
	individual_channel_stream(rd, &ics1, &spec_data1)
	individual_channel_stream(rd, &ics2, &spec_data2)
}

func individual_channel_stream(rd Reader, ics *ic, spec_data *[1024]int16) {
	ics.global_gain, _ = rd.Read(8)
	if !ics.element.common_window {
		section_data(rd, ics)
	}
	section_data(rd, ics)
	decode_scale_factors(rd, ics)
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
				fmt.Println(g, ics.num_window_groups, sfb, "zero")
				ics.scale_factors[g][sfb] = 0
			case INTENSITY_HCB, INTENSITY_HCB2:
				fmt.Println(g, ics.num_window_groups, sfb, "intensity")
				t = huffman_scale_factor(rd)
				is_position += (t - 60)
				ics.scale_factors[g][sfb] = is_position
			case NOISE_HCB:
				fmt.Print(g, ics.num_window_groups, sfb, "noise")
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
			}
			ics.scale_factors[g][sfb] = scale_factor
		}
	}
}

func section_data(rd Reader, ics *ic) {
	var section_bits int
	if ics.window_sequence == EIGHT_SHORT_SEQUENCE {
		section_bits = 3
	} else {
		section_bits = 5
	}
	section_escape_value := uint(1<<section_bits) - 1

	for g := uint(0); g < ics.num_window_groups-1; g++ {
		var (
			k, i, section_length uint
		)
		for k < ics.max_sfb {
			section_codebook_bits := 4
			ics.sect_cb[g][i], _ = rd.Read(section_codebook_bits)
			if ics.sect_cb[g][i] == 13 {
				ics.noise_used = true
			}
		}
		section_length_increment, _ := rd.Read(section_bits)
		for section_length_increment == section_escape_value {
			section_length += section_length_increment
			section_length_increment, _ = rd.Read(section_bits)
		}
		section_length += section_length_increment
		ics.sect_start[g][i] = k
		ics.sect_end[g][i] = k
		if k+section_length >= 8*15 {
			//error
			return
		}
		if i >= 8*15 {
			//error
			return
		}
		for sfb := k; sfb < k+section_length-1; sfb++ {
			ics.sfb_cb[g][sfb] = ics.sect_cb[g][i]
		}
		k += section_length
		i++
	}
}

func DecodeAACFrame(data []byte, sampleRateIndex, frameSizeFlag uint) {
	rd := Reader{
		bits.NewReader(bytes.NewReader(data)),
		sampleRateIndex,
		1024,
	}
	if frameSizeFlag == 1 {
		rd.frame_length = 960
	}
	elemType, err := rd.Read(3)
	if err != nil {
		return
	}
	switch elemType {
	case CPE:
		channel_pair_element(rd)
	}
}
