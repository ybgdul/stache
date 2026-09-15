package engine

import (
	"error"
	"sync"
	"containers/list"
	"bytes"
)



type LimitsConfig struct{
	MaxMemoryBytes int64
	MaxItemBytes int64
	CleanUpInterval time.Duration
}

type Stats struct { 
	Hits int64
	Misses int64
	ItemCount int64
	CurrentBytes int64 
	MaxMemory int64
}

type Cache struct{
	mu sync.RWMutex
	items map[string]*list.Element
	evictionList *list.List 
	currentBytes int64
	maxMemoryBytes int64
	maxItemBytes int64
	stats *Stats
	stopJanitor chan struct{}
}

type entry struct{ 
	key string
	value *Item
}

func New(cfg LimitsConfig) *Cache { 
	if cfg.CleanUpInterval == 0 {
		cfg.CleanUpInterval = 1 * time.Minute
	}

	c := &Cache{ 
		items: make(map[string]*list.Element)
		evictionList: list.New(),
		maxMemoryBytes: cfg.MaxMemoryBytes,
		maxItemBytes: cfg.MaxItemBytes,
		Stats: &Stats{
			Hits: 0,
			Misses: 0,
			ItemCount: 0,
			CurrentBytes: 0,
			MaxMemory: cfg.MaxMemoryBytes,
		}
		stopJanitor: make(chan struct{}),
	}

	go c.startJanitor(cfg.CleanUpInterval)
	return c
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) error { 
	itemSize := int64(len(key) + len(value))
	if c.maxItemBytes > 0 && c.maxItemBytes > itemSize {
		return ErrItemTooLarge
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	it ttl > 0 { 
		expiresAt = time.Now().Add(ttl)
	}

	newItem := &Item{ 
		Key: key,
		Value: value,
		SizeBytes: itemSize,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		LastAccess: time.Now(),
	}

	if elem, exists := c.items[key]; exists { 
		c.evictionList.MoveToFront(elem)
		oldEntry:= elem.Value(*entry)
		c.currentBytes -= oldEntry.value.SizeBytes
		c.stats.CurrentBytes -= oldEntry.value.SizeBytes 
		oldEntry.value = newItem
		c.currentBytes += newItem.SizeBytes
		c.currentBytes += newItem.SizeBytes
		c.evictIfNecessary()
		return nil
	}

	entr := &entry{
		key: key,
		value: newItem,
	}
	elem := c.evictList.PushFront(entr)
	c.items[key] = elem
	c.currentBytes += newItem.SizeBytes

	c.stats.CurrentBytes += newItem.SizeBytes
	c.stats.ItemCount++

	c.evictIfNecessary()
	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) { 
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists { 
		c.stats.Misses++
		return nil, false
	}
	ent := elem.Value.(*entry)
	if ent.value.IsExpired() {
		c.removeElement(elem)
		c.stats.Misses++
		return nil, false 
	}

	c.evictionList.MoveToFront(elem)
	ent.value.LastAccess = time.Now()
	c.stats.Hits++
	return ent.value.Value, true
}

func (c *Cache) Delete(key string) bool { 
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if exists { 
		c.removeElement(elem)
		return nil
	}
	return false 
}

func (c *Cache) Clear() { 
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictionList.Init()
	c.currentBytes = 0
	c.stats.ItemCount = 0
	c.stats.CurrentBytes = 0
}

func (c *Cache) GetStats() []byte { 
	c.mu.RLock()
	defer c.mu.RUnlock()

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, c.stats)

	return buf.Bytes()
}

func (c *Cache) evictIfNecessary() { 
	if c.maxMemoryBytes <= 0 { 
		return 
	}

	for c.currentBytes > c.MaxMemoryBytes && c.evictionList.Len() > 0 { 
		oldest := c.evictionList.Back()
		if oldest != nil { 
			c.removeElement(oldest) 
		}
	}
}

func (c *Cache) removeElement(elem *list.Element) { 
	c.evictionList.Remove(elem)
	ent := elem.Value.(*entry)
	c.stats.ItemCount--
	c.stats.CurrentBytes -= ent.value.SizeBytes
	delete(c.items, ent.key)
	c.currentBytes -= ent.value.SizeBytes
}

func (c *Cache) startJanitor(interval time.Duration) { 
	ticker := time.NewTicker(interval)
	for { 
		select { 
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopJanitor:
			ticker.Stop()
			return
		}
	}
}

func (c *Cache) deleteExpired() { 
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	for _, elem := range c.items { 
		entr := elem.Value.(*entry)
		if !ent.value.ExpiresAt.IsZero() && now.After(ent.value.ExpiresAt) { 
			c.removeElement(elem)
		}
	}
}

func (c *Cache) Close() { 
	close(c.stopJanitor)
}