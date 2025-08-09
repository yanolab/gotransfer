// Package transfer provides client and server for file transfer.
package transfer

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// setupTestServer starts a server for testing and returns the server address and a cleanup function.
func setupTestServer(t *testing.T) (string, func()) {
	// Create a temporary directory for the server
	serverDir, err := ioutil.TempDir("", "gotransfer-test-server")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		os.RemoveAll(serverDir)
		t.Fatalf("Failed to find available port: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	// Start the server in a goroutine
	server := NewServer(addr, serverDir)
	go func() {
		// The rpc.Register call might panic if called multiple times.
		// To avoid this, we use a sync.Once to ensure it's only called once per test run.
		// This is a workaround for the global nature of net/rpc registration.
		var once sync.Once
		register := func() {
			err := rpc.Register(&Rpc{server: server, session: &Session{mu: &sync.Mutex{}, files: make(map[SessionId]*os.File)}})
			if err != nil && err.Error() != "rpc: service already defined: Rpc" {
				t.Logf("RPC registration failed: %v", err)
			}
		}
		once.Do(register)

		if err := server.ListenAndServe(); err != nil {
			// This error is expected when the listener is closed.
			t.Logf("Server exited: %v", err)
		}
	}()
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		os.RemoveAll(serverDir)
		// We can't easily shutdown the http server gracefully,
		// but removing the directory is the main cleanup needed.
	}

	return addr, cleanup
}

func TestDownloadAndStat(t *testing.T) {
	// Create a temporary directory for the server
	serverDir, err := ioutil.TempDir("", "gotransfer-test-server")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(serverDir)

	// Create a temporary file with random content
	fileContent := make([]byte, 1024*1024) // 1MB
	if _, err := rand.Read(fileContent); err != nil {
		t.Fatalf("Failed to generate random content: %v", err)
	}
	sourceFilePath := filepath.Join(serverDir, "testfile.txt")
	if err := ioutil.WriteFile(sourceFilePath, fileContent, 0666); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	// Start the server in a goroutine
	server := NewServer(addr, serverDir)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Logf("Server exited: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Create a client and connect
	client := NewClient(addr)
	if err := client.Dial(); err != nil {
		t.Fatalf("Client failed to dial: %v", err)
	}
	defer client.Close()

	t.Run("Download", func(t *testing.T) {
		// Create a temporary directory for the client to download to
		clientDir, err := ioutil.TempDir("", "gotransfer-test-client")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(clientDir)

		// Download the file
		destFilePath := filepath.Join(clientDir, "downloaded_file.txt")
		if err := client.Download("testfile.txt", destFilePath); err != nil {
			t.Fatalf("Download failed: %v", err)
		}

		// Compare the downloaded file with the original
		downloadedContent, err := ioutil.ReadFile(destFilePath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		if !bytes.Equal(fileContent, downloadedContent) {
			t.Fatalf("Downloaded file content does not match original content")
		}
	})

	t.Run("Stat", func(t *testing.T) {
		// Stat the file
		stat, err := client.Stat("testfile.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}

		if stat.IsDir() {
			t.Errorf("expected file, got directory")
		}

		if stat.Size != int64(len(fileContent)) {
			t.Errorf("expected size %d, got %d", len(fileContent), stat.Size)
		}

		// Stat a directory
		if err := os.Mkdir(filepath.Join(serverDir, "testdir"), 0777); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		stat, err = client.Stat("testdir")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !stat.IsDir() {
			t.Errorf("expected directory, got file")
		}

		// Stat a non-existent file
		_, err = client.Stat("nonexistent")
		if err == nil {
			t.Errorf("expected error for non-existent file, got nil")
		}
	})
}

func TestSession(t *testing.T) {
	s := &Session{
		mu:    &sync.Mutex{},
		files: make(map[SessionId]*os.File),
	}

	if s.Len() != 0 {
		t.Errorf("expected 0, got %d", s.Len())
	}

	// This test is not comprehensive, as it's hard to test os.File interactions
	// without creating actual files. The main TestDownload covers the session
	// logic implicitly.
	// We can add a dummy file to test Add and Delete.
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	id := s.Add(tmpfile)
	if s.Len() != 1 {
		t.Errorf("expected 1, got %d", s.Len())
	}

	if f := s.Get(id); f != tmpfile {
		t.Errorf("expected to get the same file back")
	}

	s.Delete(id)
	if s.Len() != 0 {
		t.Errorf("expected 0, got %d", s.Len())
	}

	if f := s.Get(id); f != nil {
		t.Errorf("expected to get nil file back")
	}
}
