package server

import (
	"errors"
	"io"
	"net"
	"os"
	"stache/internal/engine"
	"stache/internal/protocol"
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

func (s *Server) Start() error { 
	if err := os.Remove(s.socketPath); err != nil && os.IsNotExist(err){ 
		return err 
	}

	l, err := net.Listen("unix", s.socketPath)
	if err != nil { 
		return err 
	}
	s.listener = l

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *Server) acceptLoop() { 
	defer s.wg.Done()

	for { 
		conn, err := s.listener.Accept()
		if err != nil { 
			select { 
			case <-s.quit: 
				return
			default: 
				continue 
			}
		}
		s.wg.Add(1)

		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleConn(c)
		}(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) { 
	defer conn.Close() 
	
	for { 
		req, err := protocol.ReadRequest(conn)
		if err != nil { 
			if errors.Is(err, io.EOF) { 
				return 
			}
			return
		}

		resp := s.dispatch(req) 
		if err := protocol.WriteResponse(conn, resp); err != nil { 
			return 
		}
	}
}

func (s *Server) dispatch(req *protocol.Request) *protocol.Response { 
	switch req.Cmd {
	case protocol.CmdGet: 
		val, found := s.cache.Get(req.Key)
		if !found {
			return &protocol.Response{Status: protocol.StatusNotFound}
		}
		return &protocol.Response{Status: protocol.StatusOk, Value: val}
	case protocol.CmdSet: 
		err := s.cache.Set(req.Key, req.Value, req.TTL); 
		if err != nil { 
			return &protocol.Response{Status: protocol.StatusErr, Err: err.Error("can't set cache")}
		}
		return &protocol.Response{Status: protocol.StatusOk}
	case protocol.CmdDelete: 
		s.cache.Delete(req.Key)
		return &protocol.Response{Status: protocol.StatusOk}
	case protocol.CmdClear: 
		s.cache.Clear()
		return &protocol.Response{Status: protocol.StatusOk}
	default:
		return &protocol.Response{Status: protocol.StatusErr, Err: "unknown command"}
	}
}

func (s *Server) Stop() error { 
	close(s.quit)
	var err error
	if s.listener != nil { 
		err = s.listener.Close()
		os.Remove(s.socketPath)
	}
	s.wg.Wait()
	return err 
}