package engine

import "time"

type Item struct { 
	Key string
	Value []byte
	SizeBytes int64
	CreatedAt time.Time
	ExpiresAt time.Time
	LastAccess time.Time
}

func (i *Item) IsExpired() bool { 
	if i.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.ExpiresAt)
}