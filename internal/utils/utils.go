package utils

func PointerTo[T any](obj T) *T {
	return &obj
}
