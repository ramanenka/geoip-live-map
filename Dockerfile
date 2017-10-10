FROM golang:1 as builder
WORKDIR $GOPATH/src/github.com/ramanenka/geoip-live-map
RUN set -ex \
  && GEOIP_ARCHIVE_URL=http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz \
  && MD5=$(wget -O - $GEOIP_ARCHIVE_URL.md5 2>/dev/null) \
  && wget -O /tmp/GeoLite2-City.tar.gz $GEOIP_ARCHIVE_URL \
  && echo "$MD5 /tmp/GeoLite2-City.tar.gz" | md5sum -c -  \
  && MMDB_FILE=$(tar -ztf /tmp/GeoLite2-City.tar.gz | grep GeoLite2-City.mmdb) \
  && tar -zxf /tmp/GeoLite2-City.tar.gz $MMDB_FILE -O > /tmp/GeoLite2-City.mmdb
COPY *.go ./
RUN go get -v -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /tmp/geoip-live-map

FROM scratch
ENV HTTP_LISTEN_ON=:80
EXPOSE $HTTP_LISTEN_ON
WORKDIR /glm
COPY index.html index.html
COPY --from=builder /tmp/geoip-live-map ./bin/
COPY --from=builder /tmp/GeoLite2-City.mmdb ./
CMD ["/glm/bin/geoip-live-map"]
