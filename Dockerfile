FROM golang:1

ENV HTTP_LISTEN_ON=:80
EXPOSE $HTTP_LISTEN_ON

RUN set -ex \
  && GEOIP_ARCHIVE_URL=http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz \
  && MD5=$(wget -O - $GEOIP_ARCHIVE_URL.md5 2>/dev/null) \
  && wget -O /tmp/GeoLite2-City.tar.gz $GEOIP_ARCHIVE_URL \
  && echo "$MD5 /tmp/GeoLite2-City.tar.gz" | md5sum -c -  \
  && MMDB_FILE=$(tar -ztf /tmp/GeoLite2-City.tar.gz | grep GeoLite2-City.mmdb) \
  && tar -zxf /tmp/GeoLite2-City.tar.gz $MMDB_FILE -O > GeoLite2-City.mmdb \
  && rm /tmp/GeoLite2-City.tar.gz

COPY index.html index.html
COPY main.go src/github.com/ramanenka/geoip-live-map/main.go
RUN (cd src/github.com/ramanenka/geoip-live-map && go get -v && go install -v)

CMD ["geoip-live-map"]