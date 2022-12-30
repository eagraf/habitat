FROM ubuntu:latest

WORKDIR /habitat

ENV HABITAT_PATH=/habitat

RUN apt-get update && \
    apt-get install -y wget make curl

RUN wget https://dist.ipfs.io/kubo/v0.14.0/kubo_v0.14.0_linux-amd64.tar.gz && tar -xvzf kubo_v0.14.0_linux-amd64.tar.gz && ./kubo/install.sh
