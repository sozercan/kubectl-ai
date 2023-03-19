package gpt3

// ToPtr converts type T to a *T as a convenience
func ToPtr[T any](i T) *T {
	return &i
}
