package provider

import (
	"strings"
	"unicode"
)

const (
	legalCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

// A name must start with a letter or underscore and
// may contain only letters, digits, underscores, and dashes.
// e.g 100_users -> _100_users
func ValidNamer(s string) string {
	if len(s) == 0 {
		return "_"
	}

	// if string is not a letter or underscore, prepend underscore
	if !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		s = "_" + s
	}

	// replace illegal characters with underscore
	for i, v := range []byte(s) {
		if !strings.Contains(legalCharacters, string(v)) {
			s = s[:i] + "_" + s[i+1:]
		}
	}

	return s
}

func ToSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	n := strings.Builder{}
	n.Grow(len(s) + 2) // nominal 2 bytes of extra space for inserted delimiters
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if vIsCap {
			v += 'a'
			v -= 'A'
		}

		if i+1 < len(s) {
			next := s[i+1]
			vIsNum := v >= '0' && v <= '9'
			nextIsCap := next >= 'A' && next <= 'Z'
			nextIsLow := next >= 'a' && next <= 'z'
			nextIsNum := next >= '0' && next <= '9'
			// add underscore if next letter case type is changed
			if (vIsCap && (nextIsLow)) || (vIsLow && (nextIsCap || nextIsNum)) || (vIsNum && (nextIsCap || nextIsLow)) {
				if vIsCap && nextIsLow {
					if prevIsCap := i > 0 && s[i-1] >= 'A' && s[i-1] <= 'Z'; prevIsCap {
						n.WriteByte('_')
					}
				}
				n.WriteByte(v)
				if vIsLow || vIsNum || nextIsNum {
					n.WriteByte('_')
				}
				continue

			}
		}

		if unicode.IsNumber(rune(v)) || unicode.IsLetter(rune(v)) {
			n.WriteByte(v)
		} else if n.Len() > 0 {
			n.WriteByte('_')
		}
	}

	return n.String()
}
