package stache

import (
	"errors"
	"net"
	"stache/internal/protocol"
	"sync"
	"time"
)

type Config struct{
	SocketPath string
	Namespace string 
	Timeout time.Duration
}

type Client struct{ 
	socketPath string 
	namespace string 
	timeout time.Duration
	sf Group
	mu sync.Mutex
	conn net.Conn
}

func NewClient(config Config) *Client { 
	if config.SocketPath == "" { 
		config.SocketPath = "/tmp/stache.sock"
	}
	if config.Timeout == 0 { 
		config.Timeout = 500 * time.Millisecond
	}
	return &Client{ 
		socketPath: config.SocketPath,
		namespace: config.Namespace,
		timeout: config.Timeout,
	}
}

func (c *Client) formatKey(key string) string {
	if c.namespace != "" { 
		return c.namespace + ":" + key;
	}
	return key
}

func (c *Client) getConn() (net.Conn, error) { 
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil { 
		return c.conn, nil
	}

	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil { 
		return nil, err
	}
	c.conn = conn 
	return c.conn, nil
}

func (c *Client) resetConn() { 
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) Get(key string) ([]byte, bool, error) {
	fullKey := c.formatKey(key)
	conn, err := c.getConn()
	if err != nil { 
		return nil, false, err 
	}

	req := &protocol.Request{
		Cmd: protocol.CmdGet,
		Key: fullKey,
	}
	if err := protocol.WriteRequest(conn, req); err != nil { 
		c.resetConn()
		return nil, false, err
	}

	resp, err := protocol.ReadResponse(conn)
	if err != nil { 
		c.resetConn()
		return nil, false, err
	}

	if resp.Status == protocol.StatusNotFound { 
		return nil, false, nil
	}
	if resp.Status == protocol.StatusErr { 
		return nil, false, errors.New(resp.Err)
	}

	return resp.Value, true, nil
}

func (c *Client) Set(key string, val []byte, ttl time.Duration) error {
	fullKey := c.formatKey(key)

	conn, err := c.getConn()
	if err != nil { 
		return err 
	}

	req := &protocol.Request{
		Cmd: protocol.CmdSet,
		Key: fullKey,
		Value: val,
	}
	if err := protocol.WriteRequest(conn, req); err != nil { 
		c.resetConn()
		return err
	}
	
	resp, err := protocol.ReadResponse(conn)
	if err != nil { 
		c.resetConn()
		return err
	}
	
	if resp.Status == protocol.StatusErr {
		return errors.New(resp.Err)
	}

	return nil
}