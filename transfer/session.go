// Package transfer provides client and server for file transfer.
package transfer

import (
	"os"
	"sync"
)

// SessionId is the type for session IDs.
type SessionId int64

// Session manages file transfer sessions.
type Session struct {
	mu      *sync.Mutex
	files   map[SessionId]*os.File
	counter SessionId
}

// Add adds a file to the session and returns a new session ID.
func (s *Session) Add(file *os.File) SessionId {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter += 1
	s.files[s.counter] = file

	return s.counter
}

// Get returns the file for a given session ID.
func (s *Session) Get(id SessionId) *os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files[id]
}

// Delete removes a session by its ID.
func (s *Session) Delete(id SessionId) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if file, exist := s.files[id]; exist {
		file.Close()
		delete(s.files, id)
	}
}

// Len returns the number of active sessions.
func (s *Session) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.files)
}
