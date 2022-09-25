FROM ubuntu:latest

WORKDIR /habitat

ENV HABITAT_PATH=/habitat

#COPY ./dist/bin /habitat/bin
#COPY ./data/procs /habitat/data/procs
#COPY ./Makefile /habitat/Makefile
#COPY ./common.mk /habitat/common.mk

RUN apt-get update && \
    apt-get install -y wget make

#RUN npm install -g serve

#CMD [ "make", "run-docker" ]
