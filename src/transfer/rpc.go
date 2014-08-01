package transfer

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Request struct {
	Id SessionId
}

type FileRequest struct {
	Filename string
}

type GetRequest struct {
	Id      SessionId
	BlockId int
}

type GetResponse struct {
	BlockId int
	Size    int64
	Data    []byte
}

type ReadRequest struct {
	Id     SessionId
	Offset int64
	Size   int
}

type ReadResponse struct {
	Size int
	Data []byte
	EOF  bool
}

type StatResponse struct {
	Type         string
	Size         int64
	LastModified time.Time
}

func (r *StatResponse) IsDir() bool {
	return r.Type == "Directory"
}

type Response struct {
	Id     SessionId
	Result bool
}

type Rpc struct {
	server  *Server
	session *Session
}

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

func (r *Rpc) Close(req Request, res *Response) error {
	r.session.Delete(req.Id)
	res.Result = true

	log.Printf("Close sessionId=%d", req.Id)

	return nil
}

func (r *Rpc) Stat(req FileRequest, res *StatResponse) error {
	path := filepath.Join(r.server.ReadDirectory, req.Filename)
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else {
		if fi.IsDir() {
			res.Type = "Directory"
		} else {
			res.Type = "File"
			res.Size = fi.Size()
		}
		res.LastModified = fi.ModTime()
	}

	log.Printf("Stat %s, %#v", req.Filename, res)

	return nil
}

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
