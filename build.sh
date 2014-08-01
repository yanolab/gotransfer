#! /bin/sh

mkdir_if_not_exists() {
    if [ ! -d $1 ]; then
        mkdir -p $1
    fi
}

initialize() {
    cwd=`dirname "${0}"`
    expr "${0}" : "/.*" > /dev/null || cwd=`(cd "${cwd}" && pwd)`

    GOPATH=${cwd}

    mkdir_if_not_exists bin/linux
    mkdir_if_not_exists bin/windows
    mkdir_if_not_exists bin/mac
}

initialize

go build -o bin/linux/gotransfer src/gotransfer.go
GOOS=windows go build -o bin/windows/gotransfer.exe src/gotransfer.go
GOOS=darwin go build -o bin/mac/gotransfer src/gotransfer.go
