# geoip-live-map
Realtime visualization of your access logs on a map

It tails an access log, searches for an IP address in each line, determines it's geo coordinates and 
animates a point on the map in the corresponding location.

![geoip-live-map](https://media.giphy.com/media/xUPGGeuxt7c3wsYODC/giphy.gif)

## Usage
```
docker run -d \
    --name geoip-live-map \
    -e LOG_FILENAME=/var/log/nginx/access.log \
    -v /var/log/nginx:/var/log/nginx:ro \
    -p 8080:80 \
    vadramanenka/geoip-live-map
```
