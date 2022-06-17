package helper

import "os"

func Getenv(name string, fallback string) string {
	value, exists := os.LookupEnv(name)
	if exists {
		return value
	}
	return fallback
}
