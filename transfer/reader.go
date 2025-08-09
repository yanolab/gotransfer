// Package transfer provides client and server for file transfer.
package transfer

import "sync"

// ReadCloser is a wrapper around the client to provide an io.ReadCloser interface.
type ReadCloser struct {
	sessionId  SessionId
	client     *Client
	mu         sync.Mutex
	blockIndex int
}

// NewReadCloser creates a new ReadCloser.
func NewReadCloser(client *Client, sessionId SessionId) *ReadCloser {
	return &ReadCloser{
		client:    client,
		sessionId: sessionId,
	}
}

// SessionId returns the session ID.
func (r *ReadCloser) SessionId() SessionId {
	return r.sessionId
}

// Read reads data from the stream.
func (r *ReadCloser) Read(buf []byte) (int, error) {
	return r.client.Read(r.sessionId, buf)
}

// Close closes the session.
func (r *ReadCloser) Close() error {
	return r.client.CloseSession(r.sessionId)
}
