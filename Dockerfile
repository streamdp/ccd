FROM golang:1.22.3-alpine as build
ARG VERSION
WORKDIR /build
ADD . /build
COPY go.* ./
RUN go get ./...

RUN apk update && apk upgrade
RUN go build -mod=readonly -o app -ldflags="-X github.com/streamdp/ccd/config.Version=$VERSION" ./

FROM alpine:3.20.1

COPY --from=build /build/app 	    /srv/app
COPY --from=build /build/site 	/srv/site/

WORKDIR /srv
EXPOSE 8080

CMD ["/srv/app"]
