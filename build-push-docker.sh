#!/bin/zsh
# compile project
docker run --platform linux/amd64 --rm -v "$PWD":/usr/src/fishnetstats -w /usr/src/fishnetstats golang:1.23.4-alpine3.21 go build -v -o stats

# build containers & push
docker buildx build --platform linux/amd64 -t nilskrau/fishnetstats:latest -f ./Dockerfile-stats .
docker push nilskrau/fishnetstats:latest
docker buildx build --platform linux/amd64 -t nilskrau/fishnet:latest -f ./Dockerfile-fishnet .
docker push nilskrau/fishnet:latest
