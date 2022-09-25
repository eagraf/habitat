FROM ubuntu:latest

WORKDIR /habitat

ENV HABITAT_PATH=/habitat

RUN apt-get update && \
    apt-get install -y wget make
