FROM alpine:latest

ARG version="default value"
ARG key1="default value"
ARG key2="default value"

WORKDIR /app

RUN echo ${version} > version.txt
RUN echo ${key1} > file-1.txt
RUN echo ${key2} > file-2.txt

ENTRYPOINT [ "cat", "/app/version.txt", "/app/file-1.txt", "/app/file-2.txt" ]
