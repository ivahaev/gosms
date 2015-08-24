FROM ubuntu:14.04
MAINTAINER Omie <intelomkar@gmail.com>

RUN apt-get update
RUN apt-get install -y build-essential
RUN apt-get install -y curl git

RUN curl -s https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -v -C /usr/local/ -xz

ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOPATH /go
ENV GOROOT /usr/local/go

RUN go get github.com/ivahaev/gosms

ADD . /go/src/github.com/ivahaev/gosms

WORKDIR /go/src/github.com/ivahaev/gosms
RUN go get

WORKDIR /go/src/github.com/ivahaev/gosms/dashboard
RUN go install github.com/ivahaev/gosms/dashboard

EXPOSE 8951

ENTRYPOINT /go/bin/dashboard

