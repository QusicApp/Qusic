package util

import "strings"

const (
	hebrewRangeMin = 1470
	hebrewRangeMax = 1524
)

func reverse[T comparable](sl []T) []T {
	rev := make([]T, len(sl))
	for i := range sl {
		rev[i] = sl[len(sl)-1-i]
	}
	return rev
}

func HebrewFix(s string) string {
	var has bool
	words := strings.Split(s, " ")
	for i, word := range words {
		if checkHeb(word) {
			has = true
			var (
				hebrewPart, englishPart string
				hebrewPartDone          bool
			)
			for _, char := range word {
				if hebrewPartDone {
					englishPart += string(char)
				} else {
					if char >= hebrewRangeMin && char <= hebrewRangeMax {
						hebrewPart += string(char)
					} else {
						hebrewPartDone = true
						englishPart += string(char)
					}
				}
			}
			words[i] = englishPart + string(reverse([]rune(hebrewPart)))
		}
	}
	if has {
		words = reverse(words)
	}
	return strings.Join(words, " ")
}

func checkHeb(s string) bool {
	r := []rune(s)
	return r[0] >= hebrewRangeMin && r[0] <= hebrewRangeMax
}
