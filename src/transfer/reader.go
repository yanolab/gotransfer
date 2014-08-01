package transfer

import "sync"

type ReadCloser struct {
	sessionId  SessionId
	client     Client
	mu         sync.Mutex
	blockIndex int
}

func (r *ReadCloser) SessionId() SessionId {
	return r.sessionId
}

func (r *ReadCloser) Read(buf []byte) (int, error) {
	return r.client.Read(r.sessionId, buf)
}

func (r *ReadCloser) Close() error {
	return r.client.Close()
}
