package transfer

import (
	"sync"
	"os"
)

type SessionId int64

type Session struct {
	mu *sync.Mutex
	files map[SessionId]*os.File
	counter SessionId
}

func (s *Session) Add(file *os.File) SessionId {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter += 1
	s.files[s.counter] = file

	return s.counter
}

func (s *Session) Get(id SessionId) *os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files[id]
}

func (s *Session) Delete(id SessionId) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if file, exist := s.files[id]; exist {
		file.Close()
		delete(s.files, id)
	}
}

func (s *Session) Len() int {
	return len(s.files)
}
