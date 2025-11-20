package formatter

import "strconv"

func Money(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	s := strconv.Itoa(n)
	b := make([]byte, 0, len(s)+len(s)/3)

	rem := len(s) % 3
	if rem > 0 {
		b = append(b, s[:rem]...)
		if len(s) > rem {
			b = append(b, ',')
		}
		s = s[rem:]
	}
	for i, ch := range s {
		if i > 0 && i%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, byte(ch))
	}

	return sign + string(b)
}
