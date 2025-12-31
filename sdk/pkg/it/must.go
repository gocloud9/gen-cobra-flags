package it

import "time"

func Must[T any](out T, err error) T {
	if err != nil {
		panic(err)
	}

	return out
}

func ParseDuration(im string) time.Duration {
	return Must(time.ParseDuration(im))
}
