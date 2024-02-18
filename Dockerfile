FROM golang:alpine AS build

RUN apk update && apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -ldflags="-w -s" -o /bin/main /app

FROM alpine

RUN apk update && apk --no-cache add ca-certificates tzdata ffmpeg python3 py3-pip py3-setuptools py3-wheel && pip3 install --upgrade youtube-dl && rm -rf /var/cache/apk/*

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/main /
COPY --from=build /app/yt-dlp /bin/

RUN chmod a+rx /bin/yt-dlp
# test yt-dlp
RUN /bin/yt-dlp --version


ENTRYPOINT ["/main"]
