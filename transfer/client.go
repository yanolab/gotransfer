// Package transfer provides client and server for file transfer.
package transfer

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
)

// Client is a file transfer client.
type Client struct {
	Addr      string
	rpcClient *rpc.Client
}

// NewClient creates a new Client.
func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

// Dial connects to the server.
func (c *Client) Dial() error {
	client, err := rpc.DialHTTP("tcp", c.Addr)
	if err != nil {
		return err
	}
	c.rpcClient = client

	return nil
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	return c.rpcClient.Close()
}

// Open opens a file on the server and returns a session ID.
func (c *Client) Open(filename string) (SessionId, error) {
	var res Response
	if err := c.rpcClient.Call("Rpc.Open", FileRequest{Filename: filename}, &res); err != nil {
		return 0, err
	}

	return res.Id, nil
}

// Stat returns file information from the server.
func (c *Client) Stat(filename string) (*StatResponse, error) {
	var res StatResponse
	if err := c.rpcClient.Call("Rpc.Stat", FileRequest{Filename: filename}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetBlock gets a block of data from the server.
func (c *Client) GetBlock(sessionId SessionId, blockId int) ([]byte, error) {
	return c.ReadAt(sessionId, int64(blockId)*BLOCK_SIZE, BLOCK_SIZE)
}

// ReadAt reads data from the server at a specific offset.
func (c *Client) ReadAt(sessionId SessionId, offset int64, size int) ([]byte, error) {
	res := &ReadResponse{Data: make([]byte, size)}
	err := c.rpcClient.Call("Rpc.ReadAt", ReadRequest{Id: sessionId, Size: size, Offset: offset}, &res)

	if res.EOF {
		err = io.EOF
	}

	if size != res.Size {
		return res.Data[:res.Size], err
	}

	return res.Data, nil
}

// Read reads data from the server.
func (c *Client) Read(sessionId SessionId, buf []byte) (int, error) {
	res := &ReadResponse{Data: buf}
	if err := c.rpcClient.Call("Rpc.Read", ReadRequest{Id: sessionId, Size: cap(buf)}, &res); err != nil {
		return 0, err
	}

	return res.Size, nil
}

// CloseSession closes a session on the server.
func (c *Client) CloseSession(sessionId SessionId) error {
	res := &Response{}
	if err := c.rpcClient.Call("Rpc.Close", Request{Id: sessionId}, &res); err != nil {
		return err
	}

	return nil
}

// Download downloads a file from the server.
func (c *Client) Download(filename, saveFile string) error {
	return c.DownloadAt(filename, saveFile, 0)
}

// DownloadAt downloads a file from the server, starting from a specific block.
func (c *Client) DownloadAt(filename, saveFile string, blockId int) error {
	stat, err := c.Stat(filename)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return errors.New(fmt.Sprintf("%s is directory.", filename))
	}

	blocks := int(stat.Size / BLOCK_SIZE)
	if stat.Size%BLOCK_SIZE != 0 {
		blocks += 1
	}

	log.Printf("Download %s in %d blocks\n", filename, blocks-blockId)

	file, err := os.OpenFile(saveFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	sessionId, err := c.Open(filename)
	if err != nil {
		return err
	}
	// Ensure session is closed even if download fails.
	defer c.CloseSession(sessionId)

	for i := blockId; i < blocks; i++ {
		buf, rerr := c.GetBlock(sessionId, i)
		if rerr != nil && rerr != io.EOF {
			return rerr
		}
		if _, werr := file.WriteAt(buf, int64(i)*BLOCK_SIZE); werr != nil {
			return werr
		}

		if (blocks-blockId) > 0 && i%((blocks-blockId)/100+1) == 0 {
			log.Printf("Downloading %s [%d/%d] blocks", filename, i-blockId+1, blocks-blockId)
		}

		if rerr == io.EOF {
			break
		}
	}
	log.Printf("Download %s completed", filename)

	return nil
}
