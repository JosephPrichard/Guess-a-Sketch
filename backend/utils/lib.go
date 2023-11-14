package utils

func Remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
