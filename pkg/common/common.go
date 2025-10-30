package common

func UseNotZero[T comparable](a, b T) T {
	var zero T
	if a == zero {
		return b
	}
	return a
}
