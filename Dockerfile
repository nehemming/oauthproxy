# builder image
FROM golang:1.16.4-alpine3.13 as builder
RUN mkdir -p /build
ADD . /build/
WORKDIR /build
RUN apk update && \
    apk add upx && \
    CGO_ENABLED=0 GOOS=linux go test ./... && \
    VERSION=$(cat ./VERSION) && \
    CGO_ENABLED=0 GOOS=linux go build -a -o oauthproxy -ldflags="-s -w -X main.Version=$VERSION" && \
    upx oauthproxy

# generate clean, final image for end users
FROM alpine:3.13
ARG REPO_USER=nehemming
ARG REPO_NAME=oauthproxy
LABEL org.opencontainers.image.source https://github.com/$REPO_USER/$REPO_NAME
COPY --from=builder /build/oauthproxy .
USER 1000:1000
# executable
ENTRYPOINT [ "./oauthproxy" ]





