package utils

import "errors"

var ErrInvalidCommand = errors.New("invalid command byte")

var (
	ErrItemTooLarge = errors.New("item size exceeds the allowed limit")
	ErrKeyNotFound = errors.New("key not found")
)