package xstrings

func SubString(txt string, start, end int) string {
	s := []rune(txt) // Return slice of UTF-8 runes

	if start >= end {
		return ""
	}
	return string(s[start:end])
}
