package utils

import "os"

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// ToPtr converts type T to a *T as a convenience.
func ToPtr[T any](i T) *T {
	return &i
}
