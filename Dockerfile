FROM alpine:3

LABEL maintainer=kuaner@gmail.com

ENV GLIBC_VERSION=2.30-r0

RUN apk add --update curl tzdata ca-certificates &&\
    curl -Lo /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub &&\
    curl -Lo glibc.apk "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk" &&\
    apk add glibc.apk &&\
    echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf &&\
    curl -Lo ffmpeg-release-amd64-static.tar.xz https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz && \
    VER=`curl https://johnvansickle.com/ffmpeg/ | grep 'release:' | tr -cd '[^0-9]*\([0-9]*\.[0-9]*\)[^0-9]*'` &&\
    tar xvJf ffmpeg-release-amd64-static.tar.xz &&\
    cp ./ffmpeg-${VER}-amd64-static/ffmpeg /usr/bin/ &&\
    cp ./ffmpeg-${VER}-amd64-static/ffprobe /usr/bin/ &&\
    update-ca-certificates &&\
    apk del curl &&\ 
    rm -rf glibc.apk /var/cache/apk/* ffmpeg-${VER}-amd64-static ffmpeg-release-amd64-static.tar.xz

ENV TZ=Asia/Shanghai