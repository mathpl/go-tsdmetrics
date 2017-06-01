package tsdmetrics

import "strings"

func CleanOpenTSDBRune(r, replace rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		fallthrough
	case r >= 'a' && r <= 'z':
		fallthrough
	case r >= '0' && r <= '9':
		fallthrough
	case r == '-' || r == '_' || r == '.' || r == '/':
		return r
	default:
		return replace
	}
}

func CleanOpenTSDB(s string) string {
	remove := func(r rune) rune {
		return CleanOpenTSDBRune(r, '_')
	}
	return strings.Map(remove, s)
}
