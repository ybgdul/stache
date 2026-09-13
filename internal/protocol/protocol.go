package protocol

import (
	"io"
	"time"
)

const (
	CmdGet byte = iota + 1
	CmdSet
	CmdDelete
	CmdClear 
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

func WriteRequest(w io.Writer, req *Request) { 

}
func ReadRequest(r io.Reader) { 

}
func WriteResponse(w io.Writer, resp *Response) { 

}
func ReadResponse(r io.Reader) { 

}