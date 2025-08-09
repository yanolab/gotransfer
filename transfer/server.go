// Package transfer provides client and server for file transfer.
package transfer

import (
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

// Server is a file transfer server.
type Server struct {
	Addr          string
	ReadDirectory string
}

// NewServer creates a new Server.
func NewServer(addr, readDirectory string) *Server {
	return &Server{Addr: addr, ReadDirectory: readDirectory}
}

// ListenAndServe starts the server.
func (srv *Server) ListenAndServe() error {
	session := &Session{mu: &sync.Mutex{}, files: make(map[SessionId]*os.File)}
	if err := rpc.Register(&Rpc{server: srv, session: session}); err != nil {
		return err
	}

	rpc.HandleHTTP()
	return http.ListenAndServe(srv.Addr, nil)
}
