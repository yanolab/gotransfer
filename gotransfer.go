package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yanolab/gotransfer/transfer"
)

var mode, addr string
var localFile, remoteFile string
var readDirectory string
var resumeId int

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
		server := transfer.NewServer(addr, readDirectory)
		server.ListenAndServe()
	default:
		client := transfer.NewClient(addr)
		if err := client.Dial(); err != nil {
			panic(err)
		}
		client.DownloadAt(remoteFile, localFile, resumeId)
	}
}
