package validations

import "unicode"

func IsPasswordValid(s string) bool {
	number := false
	lower := false
	upper := false
	for _, ch := range s {
		if unicode.IsNumber(ch) {
			number = true
		} else if unicode.IsLower(ch) {
			lower = true
		} else if unicode.IsUpper(ch) {
			upper = true
		}
	}
	return number && lower && upper && len(s) >= 5
}
