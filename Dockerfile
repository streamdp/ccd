FROM golang:1.25.1-alpine as build
ARG VERSION
ENV CGO_ENABLED=0

RUN apk update && apk upgrade

WORKDIR /build
ADD . /build
COPY go.* ./
RUN go get ./...
RUN go build -mod=readonly -o app -ldflags="-s -X github.com/streamdp/ccd/config.version=$VERSION" ./

FROM gcr.io/distroless/static-debian12

WORKDIR /srv
COPY --from=build /build/app /srv/app
COPY --from=build /build/site /srv/site/

EXPOSE 8080

CMD ["/srv/app"]