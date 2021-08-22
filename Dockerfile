FROM ubuntu:latest

WORKDIR /build

RUN apt-get update && \
    apt-get install -y wget

RUN wget https://dist.ipfs.io/go-ipfs/v0.9.0/go-ipfs_v0.9.0_linux-amd64.tar.gz && \
    tar -xvzf go-ipfs_v0.9.0_linux-amd64.tar.gz && \
    sh go-ipfs/install.sh

