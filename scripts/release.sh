#!/bin/bash
set -x

# Clean
rm -rf pkg/

# Build
GOOS=darwin GOARCH=amd64 go build -o pkg/volley-darwin-amd64/bin/volleyd ./cmd/volleyd
GOOS=darwin GOARCH=amd64 go build -o pkg/volley-darwin-amd64/bin/volleyctl ./cmd/volleyctl
GOOS=linux GOARCH=amd64 go build -o pkg/volley-linux-amd64/bin/volleyd ./cmd/volleyd
GOOS=linux GOARCH=amd64 go build -o pkg/volley-linux-amd64/bin/volleyctl ./cmd/volleyctl

# Bundle
pushd pkg
  for file in $(ls -1 .); do
    shafile="${file}.sha"
    tarfile="${file}.tar.gz"
    tar -czvf "${tarfile}" "${file}"
    shasum -a 256 "${tarfile}" > "${shafile}"
    rm -r "${file}"
  done
popd
