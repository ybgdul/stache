package stache

import (
	"fmt"
	"strings"
)

func Key(parts ...any) string { 
	strs := make([]string, len(parts))
	for i, part := range parts { 
		strs[i] = fmt.Sprintf("%v", part)
	}
	return strings.Join(strs, ":")
}