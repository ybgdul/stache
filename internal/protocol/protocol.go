package protocol

import (
	"encoding/binary"
	"io"
	"time"
)

const (
	CmdGet byte = iota + 1
	CmdSet
	CmdDelete
	CmdClear 
	CmdStats
	CmdStop
)

const ( 
	StatusOk byte = iota + 1
	StatusNotFound 
	StatusErr
)

type Request struct{
	Cmd byte
	Key string 
	Value []byte 
	TTL time.Duration
}

type Response struct { 
	Status byte
	Value []byte 
	Err string 
}

func WriteRequest(w io.Writer, req *Request) error { 
	if err := binary.Write(w, binary.BigEndian, req.Cmd); err != nil { 
		return err
	}

	keyBytes := []byte(req.Key)
	if err := binary.Write(w, binary.BigEndian, uint32(len(keyBytes))); err != nil { 
		return err
	}
	if len(keyBytes) > 0 { 
		if _, err := w.Write(keyBytes); err != nil { 
			return nil
		}
	}

	if err := binary.Write(w, binary.BigEndian, req.TTL.Nanoseconds()); err != nil { 
		return err
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(req.Value))); err != nil { 
		return err
	}

	if len(req.Value) > 0 { 
		if _, err := w.Write(req.Value); err != nil { 
			return err
		}
	}

	return nil
}
func ReadRequest(r io.Reader) (*Request, error){ 
	req := &Request{}

	if err := binary.Read(r, binary.BigEndian, &req.Cmd); err != nil { 
		return nil, err
	} 

	var keyLen uint32 
	if err := binary.Read(r, binary.BigEndian, &keyLen); err != nil { 
		return nil, err
	}
	if keyLen > 0 { 
		keyBuf := make([]byte, keyLen)
		if _, err := io.ReadFull(r, keyBuf); err != nil { 
			return nil, err
		}
		req.Key = string(keyBuf)
	}
	var ttlNano time.Duration
	if err := binary.Read(r, binary.BigEndian, &ttlNano); err != nil { 
		return nil, err
	}
	req.TTL = ttlNano

	var valLen uint32
	if err := binary.Read(r, binary.BigEndian, &valLen); err != nil { 
		return nil, err
	}
	if valLen > 0 { 
		req.Value = make([]byte, valLen)
		if _, err := io.ReadFull(r, req.Value); err != nil { 
			return nil, err
		}
	}

	return req, nil
}
func WriteResponse(w io.Writer, resp *Response) error { 
	if err := binary.Write(w, binary.BigEndian, resp.Status); err != nil {
		return err
	}

	errBytes := []byte(resp.Err)
	if err := binary.Write(w, binary.BigEndian, uint32(len(errBytes))); err != nil { 
		return err
	}
	if len(errBytes) > 0 { 
		if _, err := w.Write(errBytes); err != nil { 
			return err
		}
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(resp.Value))); err != nil { 
		return err 
	}
	if len(resp.Value) > 0 { 
		if _, err := w.Write(resp.Value); err != nil { 
			return err 
		}
	}

	return nil
}
func ReadResponse(r io.Reader) (*Response, error){ 
	resp := &Response{}
	if err := binary.Read(r, binary.BigEndian, &resp.Status ); err != nil { 
		return nil, err
	}

	var errLen uint32
	if err := binary.Read(r, binary.BigEndian, &errLen); err != nil { 
		return nil, err
	}
	if errLen > 0 { 
		errBuf := make([]byte, errLen)
		if _, err := io.ReadFull(r, errBuf); err != nil { 
			return nil, err 
		}
		resp.Err = string(errBuf)
	}

	var valLen uint32
	if err := binary.Read(r, binary.BigEndian, &valLen); err != nil { 
		return nil, err 
	}
	if valLen > 0 { 
		valBuf := make([]byte, valLen)
		if _, err := io.ReadFull(r, valBuf); err != nil { 
			return nil, err 
		}
		resp.Value = valBuf
	}

	return resp, nil
}