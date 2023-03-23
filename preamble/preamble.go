package preamble

import (
	"errors"
)

type Result[T any] struct {
	Ok T
	Err error
}

func Ok[T any](value T) *Result[T] {
	return &Result[T] {
		Ok: value,
		Err: nil,
	}
}

func Err[T any](errs ...error) *Result[T] {
	return &Result[T] {
		Ok: *new(T),
		/* this drops nil arguments automatically */
		Err: errors.Join(errs...),
	}
}

func AwaitAll[T any](channels ...<-chan T) []T {
	output := make([]T, len(channels))

	for i, channel := range channels {
		output[i] = <-channel
	}

	return output
}
