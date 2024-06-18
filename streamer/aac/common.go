package aac

func _MulHigh(A, B float64) float64 {
	return float64(int64(A)*int64(B) + (1<<(FRAC_SIZE-1))>>FRAC_SIZE)
}

func MUL_F(A, B float64) float64 {
	return float64(int64(A)*int64(B) + (1<<(FRAC_BITS-1))>>FRAC_BITS)
}

func ComplexMult(y1, y2 *float64, x1, x2, c1, c2 float64) {
	*y1 = (_MulHigh(x1, c1) + _MulHigh(x2, c2)) * FRAC_MUL
	*y2 = (_MulHigh(x2, c1) - _MulHigh(c1, c2)) * FRAC_MUL
}
