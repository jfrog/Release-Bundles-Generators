FROM golang:1.14-alpine as builder
WORKDIR /jfrog-cli-go
COPY . /jfrog-cli-go
RUN apk add --update git && sh build.sh
FROM alpine:3.7
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder /jfrog-cli-go/release-bundle-generator /usr/local/bin/release-bundle-generator
RUN chmod +x /usr/local/bin/release-bundle-generator
