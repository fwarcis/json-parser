package ascii

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}
