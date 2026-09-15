package stache

import "sync"

type call struct{ 
	wg sync.WaitGroup
	val interface{}
	err error
}

type Group struct{ 
	mu sync.Mutex
	callMap map[string]*call
}

func (g *Group) DoGroup(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.callMap == nil { 
		g.callMap = make(map[string]*call)
	}
	if c, ok := g.callMap[key]; !ok { 
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.callMap[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.callMap, key)

	return c.val, c.err
}