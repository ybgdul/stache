package server

import (
	"net"
	"stache/internal/engine"
	"sync"
)

type Server struct{
	socketPath string
	cache *engine.Cache
	listener net.Listener
	quit chan struct{}
	wg sync.WaitGroup
}

func New(socketPath string, cache *engine.Cache) *Server { 
	return &Server{
		socketPath: socketPath,
		cache: cache,
		quit: make(chan struct{}),
	}
}