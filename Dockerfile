FROM alpine:latest

RUN apk upgrade
RUN apk add openjdk17-jdk graphviz gnuplot git ruby-full make

COPY ./build/maelstrom/bb /usr/bin/bb

WORKDIR /app
COPY mlstrm .
COPY ./build/maelstrom/ .
