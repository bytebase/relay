package util

// Replace last n letter with x.
func RedactLastN(s string, n int) string {
	if n <= 0 {
		return s
	}
	out := []rune(s)
	end := 0
	if len(out)-n > 0 {
		end = len(out) - n
	}
	for i := len(out) - 1; i >= end; i-- {
		out[i] = 'x'
	}
	return string(out)
}
