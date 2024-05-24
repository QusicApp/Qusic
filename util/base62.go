package util

import (
	"fmt"
	"math/big"
	"strings"
)

const charsetInverted = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func DecodeBase62(input string) (*big.Int, error) {
	base := big.NewInt(62)
	result := big.NewInt(0)
	multiplier := big.NewInt(1)

	for i := len(input) - 1; i >= 0; i-- {
		char := string(input[i])
		index := strings.Index(charsetInverted, char)
		if index == -1 {
			return nil, fmt.Errorf("invalid character: %s", char)
		}

		indexBig := big.NewInt(int64(index))
		part := new(big.Int).Mul(indexBig, multiplier)
		result.Add(result, part)
		multiplier.Mul(multiplier, base)
	}

	return result, nil
}
