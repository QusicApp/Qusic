package aac

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
	FILL_DATA = iota + 1
	DATA_ELEMENT
	DYNAMIC_RANGE = 11
	SBR_DATA      = 12
	SBR_DATA_CRC  = 13
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

const (
	LAST_CB_IDX = 11
)
