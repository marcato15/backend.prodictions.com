FROM golang:1.8
MAINTAINER Marc Tanis "marc@blendimc.com"

WORKDIR /app

COPY . .
RUN  go get github.com/pilu/fresh

EXPOSE 1323
CMD fresh
