FROM ubuntu:latest

WORKDIR /habitat

ENV HABITAT_PATH=/habitat

COPY ./bin /habitat/bin
COPY ./data/procs /habitat/data/procs
COPY ./Makefile /habitat/Makefile
COPY ./common.mk /habitat/common.mk

RUN apt-get update && \
    apt-get install -y wget && \
    apt-get install -y make && \
    apt-get install nodejs -y && \
    apt-get install npm -y

RUN wget https://dist.ipfs.io/go-ipfs/v0.9.0/go-ipfs_v0.9.0_linux-amd64.tar.gz && \
    tar -xvzf go-ipfs_v0.9.0_linux-amd64.tar.gz && \
    sh go-ipfs/install.sh

RUN npm install -g serve

CMD [ "make", "run-docker" ]