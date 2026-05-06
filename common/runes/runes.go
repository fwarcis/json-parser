package runes

func Is(r1 rune) func(rune) bool {
	return func(r2 rune) bool {
		return r1 == r2
	}
}

func IsNot(r1 rune) func(rune) bool {
	return func(r2 rune) bool {
		return r1 != r2
	}
}
