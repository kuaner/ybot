# Start from an golang:stretch image with the latest version of Go installed
FROM golang:stretch as build-env

LABEL maintainer="kuaner@gmail.com"


WORKDIR /go/src/github.com/kuaner/ybot

# Copy the local package files to the container's workspace.
COPY . .

ENV GOPROXY=https://goproxy.io GO111MODULE=on
# Build the application using makefile

RUN make linux

# Distributed image
FROM kuaner/alpine-ffmpeg:latest

COPY --from=build-env /go/src/github.com/kuaner/ybot/build/linux/ybot /app/ybot

# Run the app by default when the container starts
CMD /app/ybot