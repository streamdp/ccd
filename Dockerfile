FROM umputun/baseimage:buildgo-master as build

WORKDIR /build
ADD . /build
COPY go.* ./
RUN go mod download

RUN apk update && apk upgrade && apk add upx
RUN go build -mod=readonly -o app ./ 
#RUN upx -9 -q app

FROM umputun/baseimage:app

LABEL org.opencontainers.image.authors="Streamdp <@streamdp>"

ENV TIME_ZONE="Europe/Minsk"

COPY --from=build /build/app 	    /srv/app
COPY --from=build /build/templates 	/srv/templates/

WORKDIR /srv

CMD ["/srv/app"]
