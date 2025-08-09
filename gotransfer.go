// gotransfer is a file transfer utility.
//
// It can be run as a server to serve files from a directory,
// or as a client to download files from a server.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yanolab/gotransfer/transfer"
)

var (
	mode          string // "client" or "server"
	addr          string // server address
	localFile     string // path to save file locally
	remoteFile    string // path of file on server
	readDirectory string // directory to serve files from
	resumeId      int    // block id to resume download from
)

func init() {
	flag.StringVar(&mode, "mode", "client", "run mode [client|server]")
	flag.StringVar(&addr, "addr", ":12427", "bind or connect addr")
	flag.StringVar(&localFile, "localfile", "", "save to localfile")
	flag.StringVar(&remoteFile, "remotefile", "", "download from remotefile")
	flag.StringVar(&readDirectory, "dir", "./", "read directory")
	flag.IntVar(&resumeId, "resumeId", 0, "resume download")
	flag.Parse()

	mode = strings.ToLower(mode)
	if mode == "client" && remoteFile == "" {
		fmt.Printf("client mode needs remotefile.\n")
		os.Exit(1)
	}
	if mode == "client" && localFile == "" {
		fmt.Println("set localfile as remotefile")
		localFile = remoteFile
	}
}

func main() {
	switch mode {
	case "server":
		fmt.Printf("Starting server on %s, serving from %s\n", addr, readDirectory)
		server := transfer.NewServer(addr, readDirectory)
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Server failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Starting client, connecting to %s\n", addr)
		client := transfer.NewClient(addr)
		if err := client.Dial(); err != nil {
			panic(err)
		}
		defer client.Close()
		if err := client.DownloadAt(remoteFile, localFile, resumeId); err != nil {
			fmt.Printf("Download failed: %v\n", err)
			os.Exit(1)
		}
	}
}
