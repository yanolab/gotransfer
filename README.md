# gotransfer

An RPC-based file transfer client and server written in Go.

## How to Build

Clone the repository and run the following command to build the binary:

```bash
go build
```

This will generate an executable file named `gotransfer`.

## How to Use

`gotransfer` operates in two modes: `client` and `server`.

### Server

To serve files, start the application in server mode.

```bash
./gotransfer -mode server -dir /path/to/serve/files -addr :12427
```

- `-mode`: Set to `server`.
- `-dir`: The directory containing the files to be served (defaults to `./`).
- `-addr`: The address and port for the server to listen on (defaults to `:12427`).

### Client

To download a file from the server, start the application in client mode.

```bash
./gotransfer -mode client -remotefile a.txt -localfile b.txt -addr 127.0.0.1:12427
```

This will download `a.txt` from the server and save it as `b.txt` on the client side.

## Command-line Flags

| Flag | Description | Default Value | Mode |
| --- | --- | --- | --- |
| `-mode` | The execution mode (`client` or `server`) | `client` | Both |
| `-addr` | The address for the server to listen on, or for the client to connect to | `:12427` | Both |
| `-dir` | The directory from which the server serves files | `./` | Server |
| `-remotefile` | The name of the remote file to download | "" | Client |
| `-localfile` | The name of the local file to save as (defaults to the value of `-remotefile`) | "" | Client |
| `-resumeId` | The session ID to resume a download | `0` | Client |
