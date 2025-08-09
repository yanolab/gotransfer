// Package transfer provides client and server for file transfer.
package transfer

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Request is a generic request with a session ID.
type Request struct {
	Id SessionId
}

// FileRequest is a request with a filename.
type FileRequest struct {
	Filename string
}

// GetRequest is a request to get a block of data.
type GetRequest struct {
	Id      SessionId
	BlockId int
}

// GetResponse is a response containing a block of data.
type GetResponse struct {
	BlockId int
	Size    int64
	Data    []byte
}

// ReadRequest is a request to read data.
type ReadRequest struct {
	Id     SessionId
	Offset int64
	Size   int
}

// ReadResponse is a response containing data read from a file.
type ReadResponse struct {
	Size int
	Data []byte
	EOF  bool
}

// StatResponse is a response containing file information.
type StatResponse struct {
	Type         string
	Size         int64
	LastModified time.Time
}

// IsDir returns true if the file is a directory.
func (r *StatResponse) IsDir() bool {
	return r.Type == "Directory"
}

// Response is a generic response.
type Response struct {
	Id     SessionId
	Result bool
}

// Rpc is the RPC service.
type Rpc struct {
	server  *Server
	session *Session
}

// Open opens a file and returns a session ID.
func (r *Rpc) Open(req FileRequest, res *Response) error {
	path := filepath.Join(r.server.ReadDirectory, req.Filename)
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	res.Id = r.session.Add(file)
	res.Result = true

	log.Printf("Open %s, sessionId=%d", req.Filename, res.Id)

	return nil
}

// Close closes a session.
func (r *Rpc) Close(req Request, res *Response) error {
	r.session.Delete(req.Id)
	res.Result = true

	log.Printf("Close sessionId=%d", req.Id)

	return nil
}

// Stat returns file information.
func (r *Rpc) Stat(req FileRequest, res *StatResponse) error {
	path := filepath.Join(r.server.ReadDirectory, req.Filename)
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	if err != nil {
		return err
	}

	if fi.IsDir() {
		res.Type = "Directory"
	} else {
		res.Type = "File"
		res.Size = fi.Size()
	}
	res.LastModified = fi.ModTime()

	log.Printf("Stat %s, %#v", req.Filename, res)

	return nil
}

// ReadAt reads data from a file at a specific offset.
func (r *Rpc) ReadAt(req ReadRequest, res *ReadResponse) error {
	file := r.session.Get(req.Id)
	if file == nil {
		return errors.New("You must call open first.")
	}

	res.Data = make([]byte, req.Size)
	n, err := file.ReadAt(res.Data, req.Offset)
	if err != nil && err != io.EOF {
		return err
	}

	if err == io.EOF {
		res.EOF = true
	}

	res.Size = n
	res.Data = res.Data[:n]

	log.Printf("ReadAt sessionId=%d, Offset=%d, n=%d", req.Id, req.Offset, res.Size)

	return nil
}

// Read reads data from a file.
func (r *Rpc) Read(req ReadRequest, res *ReadResponse) error {
	file := r.session.Get(req.Id)
	if file == nil {
		return errors.New("You must call open first.")
	}

	res.Data = make([]byte, req.Size)
	n, err := file.Read(res.Data)
	if err != nil && err != io.EOF {
		return err
	}

	if err == io.EOF {
		res.EOF = true
	}

	res.Size = n
	res.Data = res.Data[:res.Size]

	log.Printf("Read sessionId=%d, read=%d[bytes]", req.Id, res.Size)

	return nil
}
