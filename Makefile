NAME=ybot
PKG=github.com/kuaner/ybot
VERSION=git-$(shell git describe --always --dirty)
IMAGE_TAG=$(VERSION)
IAMGE_REPO=kuaner


linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -a -tags netgo -installsuffix netgo -installsuffix cgo -ldflags '-w -s' -ldflags "-X main.Version=$(VERSION)" \
		-o ./build/linux/ybot $(PKG)

push:
	docker build -t ${IAMGE_REPO}/ybot:${IMAGE_TAG} .
	docker push ${IAMGE_REPO}/ybot:${IMAGE_TAG}
	docker rmi ${IAMGE_REPO}/ybot:${IMAGE_TAG}