FROM ubuntu:22.04

RUN apt-get update && apt-get upgrade
RUN apt-get install -y openjdk-17-jdk graphviz gnuplot git ruby-full

COPY ./bb /usr/bin/bb

WORKDIR /maelstrom
COPY ./maelstrom .
COPY ./Makefile .

CMD bash

