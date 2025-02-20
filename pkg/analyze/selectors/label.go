package selectors

type labeled[T any] struct {
	Label string
	Value T
}
